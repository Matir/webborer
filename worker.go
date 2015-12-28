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

package main

import (
	"fmt"
	"golang.org/x/net/html"
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
	Handle(*url.URL, io.Reader)
}

// Workers do the work of connecting to the server, issuing the request, and
// then optionally parsing the response.  Normally a pool of several workers
// will be used due to network latency.
type Worker struct {
	// client for connections
	client *http.Client
	// Channel for URLs to scan
	src <-chan *url.URL
	// Function to add future work
	adder QueueAddFunc
	// Function to mark work done
	done QueueDoneFunc
	// Channel for scan results
	results chan<- Result
	// Settings
	settings *ScanSettings
	// HTML worker to parse page
	pageWorker PageWorker
	// Channel to trigger stopping
	stop chan bool
	// Request for redirection
	redir *http.Request
}

type HTMLWorker struct {
	// Function to add future work
	adder QueueAddFunc
}

// Construct a worker with given settings.
func NewWorker(settings *ScanSettings,
	factory ClientFactory,
	src <-chan *url.URL,
	adder QueueAddFunc,
	done QueueDoneFunc,
	results chan<- Result) *Worker {
	w := &Worker{
		client:   factory.Get(),
		settings: settings,
		src:      src,
		adder:    adder,
		done:     done,
		results:  results,
		stop:     make(chan bool),
	}

	// Install redirect handler
	redirHandler := func(req *http.Request, _ []*http.Request) error {
		w.redir = req
		return fmt.Errorf("Stop redirect.")
	}
	w.client.CheckRedirect = redirHandler

	return w
}

func (w *Worker) SetPageWorker(pw PageWorker) {
	w.pageWorker = pw
}

func (w *Worker) Run() {
	for true {
		select {
		case <-w.stop:
			return
		case task, ok := <-w.src:
			if !ok {
				return
			}
			w.HandleURL(task)
		}
	}
}

func (w *Worker) RunInBackground() {
	go w.Run()
}

func (w *Worker) Stop() {
	w.stop <- true
}

func (w *Worker) HandleURL(task *url.URL) {
	Logf(LogDebug, "Trying Raw URL (unmangled): %s", task.String())
	w.TryURL(task)
	if !URLIsDir(task) {
		w.TryMangleURL(task)
		for _, ext := range w.settings.Extensions {
			task := *task
			task.Path += "." + ext
			w.TryURL(&task)
			w.TryMangleURL(&task)
		}
	}
	// Mark as done
	w.done(1)
}

func (w *Worker) TryMangleURL(task *url.URL) {
	if !w.settings.Mangle {
		return
	}
	clone := *task
	spos := strings.LastIndex(clone.Path, "/")
	if spos == -1 {
		return
	}
	dirname := clone.Path[:spos]
	basename := clone.Path[spos+1:]
	for _, newname := range Mangle(basename) {
		clone := clone
		clone.Path = dirname + "/" + newname
		w.TryURL(&clone)
	}
}

func (w *Worker) TryURL(task *url.URL) {
	Logf(LogInfo, "Trying: %s", task.String())
	w.redir = nil
	req := w.MakeRequest(task)
	if resp, err := w.client.Do(req); err != nil && w.redir == nil {
		result := Result{URL: task, Error: err}
		if resp != nil {
			result.Code = resp.StatusCode
		}
		w.results <- result
	} else {
		defer resp.Body.Close()
		// Do we keep going?
		if URLIsDir(task) && KeepSpidering(resp.StatusCode) {
			Logf(LogDebug, "Referring %s back for spidering.", task.String())
			w.adder(task)
		}
		if w.pageWorker != nil && w.pageWorker.Eligible(resp) {
			w.pageWorker.Handle(task, resp.Body)
		}
		var redir *url.URL
		if w.redir != nil {
			redir = w.redir.URL
		}
		var length int64
		if resp.ContentLength > 0 {
			length = resp.ContentLength
		}
		w.results <- Result{
			URL:    task,
			Code:   resp.StatusCode,
			Redir:  redir,
			Length: length,
		}
	}
	if w.settings.SleepTime != 0 {
		time.Sleep(w.settings.SleepTime)
	}
}

func (w *Worker) MakeRequest(task *url.URL) *http.Request {
	// TODO: support other methods
	req, _ := http.NewRequest("GET", task.String(), nil)
	req.Header.Set("User-Agent", w.settings.UserAgent)
	return req
}

func NewHTMLWorker(adder QueueAddFunc) *HTMLWorker {
	return &HTMLWorker{adder: adder}
}

func (w *HTMLWorker) Handle(URL *url.URL, body io.Reader) {
	links := w.GetLinks(body)
	foundURLs := make([]*url.URL, 0, len(links))
	for _, l := range links {
		u, err := url.Parse(l)
		if err != nil {
			Logf(LogInfo, "Error parsing URL (%s): %s", l, err.Error())
			continue
		}
		foundURLs = append(foundURLs, URL.ResolveReference(u))
	}
	w.adder(foundURLs...)
}

func (*HTMLWorker) Eligible(resp *http.Response) bool {
	ct := resp.Header.Get("Content-type")
	if strings.ToLower(ct) == "text/html" {
		return false
	}
	return resp.ContentLength > 0 && resp.ContentLength < 1024*1024
}

func (*HTMLWorker) GetLinks(body io.Reader) []string {
	tree, err := html.Parse(body)
	if err != nil {
		Logf(LogInfo, "Unable to parse HTML document: %s", err.Error())
		return nil
	}
	links := make([]string, 0)
	var handleNode func(*html.Node)
	handleNode = func(node *html.Node) {
		if node.Type == html.ElementNode {
			if strings.ToLower(node.Data) == "a" {
				for _, a := range node.Attr {
					if strings.ToLower(a.Key) == "href" {
						links = append(links, a.Val)
						break
					}
				}
			}
		}
		// Handle children
		for n := node.FirstChild; n != nil; n = n.NextSibling {
			handleNode(n)
		}
	}
	handleNode(tree)
	return DedupeStrings(links)
}

// Starts a batch of workers based on the relevant settings.
func StartWorkers(settings *ScanSettings,
	factory ClientFactory,
	src <-chan *url.URL,
	adder QueueAddFunc,
	done QueueDoneFunc,
	results chan<- Result) []*Worker {
	count := settings.Workers
	workers := make([]*Worker, count)
	for i := 0; i < count; i++ {
		workers[i] = NewWorker(settings, factory, src, adder, done, results)
		workers[i].RunInBackground()
	}
	return workers
}

// Set PageWorker for a bunch of Workers
func SetPageWorkers(batch []*Worker, pw PageWorker) {
	for _, w := range batch {
		w.SetPageWorker(pw)
	}
}

// Mangle a basename
func Mangle(basename string) []string {
	mangleRules := []string{
		".%s.swp", // VIM Swap File
		"%s~",     // Backup file
		"%s.bak",  // Backup file
		"%s.orig", //Backup file
	}
	results := make([]string, len(mangleRules))
	for i, rule := range mangleRules {
		results[i] = fmt.Sprintf(rule, basename)
	}
	return results
}
