// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package worker provides the workers that do the actual loading & processing
// of pages.
package worker

import (
	"fmt"
	"github.com/Matir/webborer/client"
	"github.com/Matir/webborer/logging"
	"github.com/Matir/webborer/results"
	ss "github.com/Matir/webborer/settings"
	"github.com/Matir/webborer/task"
	"github.com/Matir/webborer/util"
	"github.com/Matir/webborer/workqueue"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Stoppable interface {
	Stop()
}

type PageWorker interface {
	Eligible(*http.Response) bool
	Handle(*task.Task, io.Reader)
}

// Workers do the work of connecting to the server, issuing the request, and
// then optionally parsing the response.  Normally a pool of several workers
// will be used due to network latency.
type Worker struct {
	// client for connections
	client client.Client
	// Channel for URLs to scan
	src <-chan *task.Task
	// Function to add future work
	adder workqueue.QueueAddFunc
	// Function to mark work done
	done workqueue.QueueDoneFunc
	// Channel for scan results
	rchan chan<- *results.Result
	// Settings
	settings *ss.ScanSettings
	// HTML worker to parse page
	pageWorker PageWorker
	// Channel to trigger stopping
	stop chan bool
	// Request for redirection
	redir *http.Request
	// Channel to signal worker stopping
	waitq chan bool
}

// Construct a worker with given settings.
func NewWorker(settings *ss.ScanSettings,
	factory client.ClientFactory,
	src <-chan *task.Task,
	adder workqueue.QueueAddFunc,
	done workqueue.QueueDoneFunc,
	rchan chan<- *results.Result) *Worker {
	w := &Worker{
		client:   factory.Get(),
		settings: settings,
		src:      src,
		adder:    adder,
		done:     done,
		rchan:    rchan,
		stop:     make(chan bool),
		waitq:    make(chan bool),
	}

	// Install redirect handler
	redirHandler := func(req *http.Request, _ []*http.Request) error {
		w.redir = req
		return fmt.Errorf("Stop redirect.")
	}
	w.client.SetCheckRedirect(redirHandler)

	return w
}

func (w *Worker) SetPageWorker(pw PageWorker) {
	w.pageWorker = pw
}

// Run the worker, processing input from a channel until either signalled to
// stop or the input channel is closed.
func (w *Worker) Run() {
	defer func() {
		w.waitq <- true
	}()
	for true {
		select {
		case <-w.stop:
			return
		case task, ok := <-w.src:
			if !ok { // channel closed
				return
			}
			w.HandleTask(task)
		}
	}
}

func (w *Worker) RunInBackground() {
	go w.Run()
}

func (w *Worker) Stop() {
	w.stop <- true
}

func (w *Worker) Wait() {
	<-w.waitq
}

func (w *Worker) HandleTask(task *task.Task) {
	logging.Logf(logging.LogDebug, "Trying Raw URL (unmangled): %s", task.String())
	withMangle := w.TryTask(task)
	if !util.URLIsDir(task.URL) {
		if withMangle {
			w.TryMangleTask(task)
		}
		if !util.URLHasExtension(task.URL) {
			for _, ext := range w.settings.Extensions {
				task := task.Copy()
				task.URL.Path += "." + ext
				if w.TryTask(task) {
					w.TryMangleTask(task)
				}
			}
		}
	}
	// Mark as done
	w.done(1)
}

func (w *Worker) TryMangleTask(task *task.Task) {
	if !w.settings.Mangle {
		return
	}
	clone := task.Copy()
	spos := strings.LastIndex(clone.URL.Path, "/")
	if spos == -1 {
		return
	}
	dirname := clone.URL.Path[:spos]
	basename := clone.URL.Path[spos+1:]
	for _, newname := range Mangle(basename) {
		clone := clone.Copy()
		clone.URL.Path = dirname + "/" + newname
		w.TryTask(clone)
	}
}

func (w *Worker) TryTask(task *task.Task) bool {
	logging.Logf(logging.LogInfo, "Trying: %s", task.String())
	tryMangle := false
	w.redir = nil
	// TODO: handle Host & Headers!
	if resp, err := w.client.Request(task.URL, task.Host, task.Header); err != nil && w.redir == nil {
		// TODO: add host, headers, group, etc.
		result := &results.Result{URL: task.URL, Error: err}
		if resp != nil {
			result.Code = resp.StatusCode
		}
		w.rchan <- result
	} else {
		defer resp.Body.Close()
		// Do we keep going?
		if util.URLIsDir(task.URL) && w.KeepSpidering(resp.StatusCode) {
			logging.Logf(logging.LogDebug, "Referring %s back for spidering.", task.String())
			w.adder(task)
		}
		if w.redir != nil {
			logging.Logf(logging.LogDebug, "Referring redirect %s back.", w.redir.URL.String())
			t := task.Copy()
			t.URL = w.redir.URL
			w.adder(t)
		}
		if w.pageWorker != nil && w.pageWorker.Eligible(resp) {
			w.pageWorker.Handle(task, resp.Body)
		}
		var redir *url.URL
		if w.redir != nil {
			redir = w.redir.URL
		}
		w.rchan <- &results.Result{
			URL:         task.URL,
			Host:        task.Host,
			Code:        resp.StatusCode,
			Redir:       redir,
			Length:      resp.ContentLength,
			ContentType: resp.Header.Get("Content-Type"),
		}
		tryMangle = w.KeepSpidering(resp.StatusCode)
	}
	if w.settings.SleepTime != 0 {
		time.Sleep(w.settings.SleepTime)
	}
	return tryMangle
}

// Should we keep spidering from this code?
func (w *Worker) KeepSpidering(code int) bool {
	if w.settings.RunMode == ss.RunModeDotProduct {
		return false
	}
	for _, v := range w.settings.SpiderCodes {
		if code == v {
			return true
		}
	}
	return false
}

// Starts a batch of workers based on the relevant settings.
func StartWorkers(settings *ss.ScanSettings,
	factory client.ClientFactory,
	src <-chan *task.Task,
	adder workqueue.QueueAddFunc,
	done workqueue.QueueDoneFunc,
	rchan chan<- *results.Result) []*Worker {
	count := settings.Workers
	workers := make([]*Worker, count)
	for i := 0; i < count; i++ {
		workers[i] = NewWorker(settings, factory, src, adder, done, rchan)
		workers[i].RunInBackground()
		if settings.ParseHTML {
			workers[i].SetPageWorker(NewHTMLWorker(adder))
		}
	}
	return workers
}

// Mangle a basename
func Mangle(basename string) []string {
	mangleRules := []string{
		".%s.swp", // VIM Swap File
		"%s~",     // Backup file
		"%s.bak",  // Backup file
		"%s.orig", // Backup file
	}
	res := make([]string, len(mangleRules))
	for i, rule := range mangleRules {
		res[i] = fmt.Sprintf(rule, basename)
	}
	return res
}
