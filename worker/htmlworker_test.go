// Copyright 2017 Google Inc. All Rights Reserved.
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
	"github.com/Matir/webborer/results"
	"github.com/Matir/webborer/task"
	"net/url"
	"strings"
	"testing"
)

func compareURLSlice(a, b []*url.URL) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].String() != b[i].String() {
			return false
		}
	}
	return true
}

var smallHTMLDoc = `
<html>
<body>
<a href='link1'>Link1</a>
<a href='http://www.example.org/'>Ex</a>
<a href="/link2/page">page</a>
<img src='/img/x.png'>
</body>
</html>`

func TestHandle(t *testing.T) {
	// Setup environment
	resultlist := make([]*task.Task, 0)
	adder := func(f ...*task.Task) {
		resultlist = append(resultlist, f...)
	}
	htmlWorker := NewHTMLWorker(adder)
	base, err := url.Parse("http://www.example.com/subdir/")
	if err != nil {
		t.Fatalf("Error in parsing base url: %v", err)
	}

	// Run the worker
	madeTask := task.NewTaskFromURL(base)
	htmlWorker.Handle(madeTask, strings.NewReader(smallHTMLDoc), results.NewResultForTask(madeTask))

	// Make slice of expected URL
	expected := make([]*url.URL, 0)
	expectedStr := []string{
		"http://www.example.com/subdir/link1",
		"http://www.example.com/subdir",
		"http://www.example.org/",
		"http://www.example.com/link2/page",
		"http://www.example.com/link2",
		"http://www.example.com/img/x.png",
		"http://www.example.com/img",
	}
	for i := range expectedStr {
		u, err := url.Parse(expectedStr[i])
		if err != nil {
			t.Fatalf("Error in parsing expected url: %v", err)
		}
		expected = append(expected, u)
	}

	// Tests
	uResults := make([]*url.URL, 0, len(resultlist))
	for _, v := range resultlist {
		uResults = append(uResults, v.URL)
	}
	if !compareURLSlice(expected, uResults) {
		t.Fatalf("Results do not match.  Expected: %v, got %v.", expected, resultlist)
	}
}
