// Copyright 2016 Google Inc. All Rights Reserved.
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

package worker

import (
	"github.com/Matir/webborer/client/mock"
	"github.com/Matir/webborer/results"
	"github.com/Matir/webborer/settings"
	"github.com/Matir/webborer/task"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func noopInt(_ int)           {}
func noopUrl(_ ...*task.Task) {}

func TestNewWorker(t *testing.T) {
	ss := &settings.ScanSettings{}
	src := make(chan *task.Task)
	rchan := make(chan *results.Result)
	worker := NewWorker(ss, &mock.MockClientFactory{}, src, noopUrl, noopInt, rchan)
	if worker == nil {
		t.Fatal("Expected to receive a worker, got nil!")
	}
}

func TryTaskHelper(u *task.Task, resp *http.Response) *Worker {
	client := &mock.MockClient{}
	if resp != nil {
		client.NextResponse = resp
	}
	ss := &settings.ScanSettings{
		SpiderCodes: []int{200},
	}
	rchan := make(chan *results.Result)
	w := &Worker{
		client:   client,
		settings: ss,
		rchan:    rchan,
		adder:    noopUrl,
	}
	defer close(rchan)
	go func() {
		for range rchan {
		}
	}()
	w.TryTask(u)
	return w
}

func TestTryURL_Basic(t *testing.T) {
	resp := mock.ResponseFromString("")
	resp.StatusCode = 200
	u := task.NewTaskFromURL(&url.URL{Scheme: "http", Host: "localhost", Path: "/"})
	TryTaskHelper(u, resp)
	// TODO: check which requests were made
}

func TestTryURL_Error(t *testing.T) {
	u := task.NewTaskFromURL(&url.URL{Scheme: "http", Host: "localhost", Path: "/"})
	TryTaskHelper(u, nil)
	// TODO: check which requests were made
}

func TestTryMangleURL_Basic(t *testing.T) {
	resp := mock.ResponseFromString("")
	resp.StatusCode = 200
	client := &mock.MockClient{
		ForeverResponse: resp,
	}
	ss := &settings.ScanSettings{
		SpiderCodes: []int{200},
		Mangle:      true,
	}
	rchan := make(chan *results.Result)
	defer close(rchan)
	go func() {
		for range rchan {
		}
	}()
	w := &Worker{
		client:   client,
		settings: ss,
		rchan:    rchan,
		adder:    noopUrl,
	}
	u := &url.URL{Scheme: "http", Host: "localhost", Path: "/"}
	w.TryMangleTask(task.NewTaskFromURL(u))
	// TODO: check which requests were made
}

func TestTryHandleURL_Basic(t *testing.T) {
	resp := mock.ResponseFromString("")
	resp.StatusCode = 200
	client := &mock.MockClient{
		ForeverResponse: resp,
	}
	ss := &settings.ScanSettings{
		SpiderCodes: []int{200},
		Mangle:      true,
		Extensions:  []string{"html", "php"},
	}
	rchan := make(chan *results.Result)
	defer close(rchan)
	go func() {
		for range rchan {
		}
	}()
	w := &Worker{
		client:   client,
		settings: ss,
		rchan:    rchan,
		adder:    noopUrl,
		done:     noopInt,
	}
	u := &url.URL{Scheme: "http", Host: "localhost", Path: "/index"}
	w.HandleTask(task.NewTaskFromURL(u))
	// TODO: check which requests were made
}

func TestStartWorkers_SingleIteration(t *testing.T) {
	ss := &settings.ScanSettings{
		Workers:   2,
		ParseHTML: true,
	}
	schan := make(chan *task.Task)
	rchan := make(chan *results.Result)
	u, _ := url.Parse("http://www.example.com")
	for i, w := range StartWorkers(
		ss,
		&mock.MockClientFactory{},
		schan,
		noopUrl,
		noopInt,
		rchan) {
		// Send the input
		schan <- task.NewTaskFromURL(u)
		// Read the result
		<-rchan
		// Both methods of signalling closure
		if i%2 == 0 {
			w.Stop()
		} else {
			close(schan)
		}
		w.Wait()
	}
}

func TestMangle(t *testing.T) {
	foo := "foo"
	for _, r := range Mangle(foo) {
		if !strings.Contains(r, foo) {
			t.Errorf("Expected %s within %s", foo, r)
		}
	}
}

type FakePageWorker struct{}

func (*FakePageWorker) Eligible(_ *http.Response) bool {
	return true
}

func (*FakePageWorker) Handle(_ *task.Task, _ io.Reader, _ *results.Result) {}

func TestSetPageWorker(t *testing.T) {
	w := &Worker{}
	pw := &FakePageWorker{}
	w.SetPageWorker(pw)
	if w.pageWorker != pw {
		t.Fatalf("Pageworker not properly set.")
	}
}
