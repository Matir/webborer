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
	"github.com/Matir/gobuster/client/mock"
	"github.com/Matir/gobuster/results"
	"github.com/Matir/gobuster/settings"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func noopInt(_ int)         {}
func noopUrl(_ ...*url.URL) {}

func TestNewWorker(t *testing.T) {
	ss := &settings.ScanSettings{}
	src := make(chan *url.URL)
	rchan := make(chan results.Result)
	worker := NewWorker(ss, &mock.MockClientFactory{}, src, noopUrl, noopInt, rchan)
	if worker == nil {
		t.Fatal("Expected to receive a worker, got nil!")
	}
}

func TryURLHelper(u *url.URL, resp *http.Response) *Worker {
	client := &mock.MockClient{}
	if resp != nil {
		client.NextResponse = resp
	}
	ss := &settings.ScanSettings{
		SpiderCodes: []int{200},
	}
	rchan := make(chan results.Result)
	w := &Worker{
		client:   client,
		settings: ss,
		rchan:    rchan,
		adder:    noopUrl,
	}
	go func() {
		for range rchan {
		}
	}()
	w.TryURL(u)
	return w
}

func TestTryURL_Basic(t *testing.T) {
	resp := mock.ResponseFromString("")
	resp.StatusCode = 200
	u := &url.URL{Scheme: "http", Host: "localhost", Path: "/"}
	TryURLHelper(u, resp)
}

func TestTryURL_Error(t *testing.T) {
	u := &url.URL{Scheme: "http", Host: "localhost", Path: "/"}
	TryURLHelper(u, nil)
}

func TestStartWorkers_Single(t *testing.T) {
	ss := &settings.ScanSettings{
		Workers: 1,
	}
	schan := make(chan *url.URL)
	rchan := make(chan results.Result)
	for _, w := range StartWorkers(
		ss,
		&mock.MockClientFactory{},
		schan,
		noopUrl,
		noopInt,
		rchan) {
		w.Stop()
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
