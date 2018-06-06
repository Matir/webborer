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
	"github.com/matir/webborer/client"
	"github.com/matir/webborer/logging"
	"github.com/matir/webborer/results"
	ss "github.com/matir/webborer/settings"
	"github.com/matir/webborer/task"
	"github.com/matir/webborer/util"
	"github.com/matir/webborer/workqueue"
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
		case t, ok := <-w.src:
			if !ok { // channel closed
				return
			}
			w.HandleTask(t)
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

func (w *Worker) HandleTask(t *task.Task) {
	logging.Logf(logging.LogDebug, "Trying Raw URL (unmangled): %s", t.String())
	withMangle := w.TryTask(t)
	if !util.URLIsDir(t.URL) {
		if withMangle {
			w.TryMangleTask(t)
		}
		if !util.URLHasExtension(t.URL) {
			for _, ext := range w.settings.Extensions {
				t := t.Copy()
				t.URL.Path += "." + ext
				if w.TryTask(t) {
					w.TryMangleTask(t)
				}
			}
		}
	}
	// Mark as done
	w.done(1)
}

func (w *Worker) TryMangleTask(t *task.Task) {
	if !w.settings.Mangle {
		return
	}
	clone := t.Copy()
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

func (w *Worker) TryTask(t *task.Task) bool {
	logging.Logf(logging.LogInfo, "Trying: %s", t.String())
	tryMangle := false
	w.redir = nil
	if resp, err := w.client.Request(t.URL, t.Host, t.Header); err != nil && w.redir == nil {
		// TODO: add host, headers, group, etc.
		result := results.NewResult(t.URL, t.Host)
		result.Error = err
		if resp != nil {
			result.Code = resp.StatusCode
		}
		w.rchan <- result
	} else {
		defer resp.Body.Close()
		// Do we keep going?
		if util.URLIsDir(t.URL) && w.KeepSpidering(resp.StatusCode) {
			logging.Logf(logging.LogDebug, "Referring %s back for spidering.", t.String())
			w.adder(t)
		}
		if w.redir != nil {
			logging.Logf(logging.LogDebug, "Referring redirect %s back.", w.redir.URL.String())
			t := t.Copy()
			t.URL = w.redir.URL
			w.adder(t)
		}
		if w.pageWorker != nil && w.pageWorker.Eligible(resp) {
			w.pageWorker.Handle(t, resp.Body)
		}
		var redir *url.URL
		if w.redir != nil {
			redir = w.redir.URL
		}
		res := results.NewResult(t.URL, t.Host)
		res.Code = resp.StatusCode
		res.Redir = redir
		res.Length = resp.ContentLength
		res.ContentType = resp.Header.Get("Content-Type")
		w.rchan <- res
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
		if settings.ParseHTML && settings.RunMode == ss.RunModeEnumeration {
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
