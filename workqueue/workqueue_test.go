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

package workqueue

import (
	"fmt"
	"net/url"
	"strconv"
	"testing"
)

func TestWorkqueue_Basic(t *testing.T) {
	filter := func(_ *url.URL) bool { return true }

	queue := NewWorkQueue(5, nil, false)
	queue.filter = filter
	queue.RunInBackground()
	for i := 0; i < 20; i++ {
		s := fmt.Sprintf("%d", i)
		u := &url.URL{Path: s}
		queue.AddURLs(u)
	}
	queue.InputFinished()
	out := queue.GetWorkChan()
	for i := 0; i < 20; i++ {
		o, ok := <-out
		s := fmt.Sprintf("%d", i)
		if !ok {
			t.Errorf("Was expecting %s, got error!", s)
			break
		}
		if o.Path != s {
			t.Errorf("Out of order responses, got %s, expected %s", o.Path, s)
		}
		queue.ctr.Done(1)
	}
	queue.WaitPipe()
}

func TestWorkqueue_Reject(t *testing.T) {
	filter := func(_ *url.URL) bool { return false }

	queue := NewWorkQueue(5, nil, false)
	queue.filter = filter
	queue.RunInBackground()
	for i := 0; i < 20; i++ {
		s := fmt.Sprintf("%d", i)
		u := &url.URL{Path: s}
		queue.AddURLs(u)
	}
	queue.InputFinished()
	out := queue.GetWorkChan()
	var i int
	for range out {
		i++
		queue.GetDoneFunc()(1)
	}
	queue.WaitPipe()
	if i > 0 {
		t.Errorf("Expecting all URLs to be filtered, got output!")
	}
}

func TestWorkqueue_PartialReject(t *testing.T) {
	rounds := 20
	filter := func(u *url.URL) bool {
		i, _ := strconv.Atoi(u.Path)
		return i < (rounds / 2)
	}

	queue := NewWorkQueue(5, nil, false)
	queue.peek()
	queue.filter = filter
	queue.RunInBackground()
	for i := 0; i < rounds; i++ {
		s := fmt.Sprintf("%d", i)
		u := &url.URL{Path: s}
		queue.GetAddFunc()(u)
	}
	queue.InputFinished()
	out := queue.GetWorkChan()
	var i int
	for range out {
		i++
		queue.GetDoneFunc()(1)
	}
	queue.WaitPipe()
	if i != (rounds / 2) {
		t.Errorf("Expecting some URLs to be filtered, got %d vs %d output!", i, rounds/2)
	}
}

func TestWorkqueue_Funcs(_ *testing.T) {
	queue := NewWorkQueue(5, nil, false)
	queue.GetAddFunc()
	queue.GetAddCount()
	queue.GetDoneFunc()
}

func TestMakeScopeFunc(t *testing.T) {
	// TODO: test multuple bases
	urlParse := func(s string) *url.URL {
		u, _ := url.Parse(s)
		return u
	}
	baseURL, _ := url.Parse("http://localhost/foo")
	results := []struct {
		u       *url.URL
		basic   bool
		upgrade bool
	}{
		{urlParse("http://localhost/foo/bar"), true, true},
		{urlParse("http://localhost/bar"), false, false},
		{urlParse("https://localhost/foo/bar"), false, true},
		{urlParse("https://localhost/bar"), false, false},
		{urlParse("https://localhost/"), false, false},
		{urlParse("http://localhost/foo"), true, true},
		{urlParse("https://localhost/foo"), false, true},
	}

	withoutUpgrade := makeScopeFunc([]*url.URL{baseURL}, false)
	withUpgrade := makeScopeFunc([]*url.URL{baseURL}, true)
	for _, res := range results {
		if withoutUpgrade(res.u) != res.basic {
			t.Errorf("URL %v did not give expected result: %v", res.u, res.basic)
		}
		if withUpgrade(res.u) != res.upgrade {
			t.Errorf("URL %v did not give expected result with upgrade: %v", res.u, res.upgrade)
		}
	}
}
