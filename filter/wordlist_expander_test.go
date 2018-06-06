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

package filter

import (
	"github.com/matir/webborer/task"
	"net/url"
	"testing"
)

func TestProcessWordlist(t *testing.T) {
	wl := []string{"a", "b/", "c.txt"}
	expected := []string{"a", "a/", "b/", "c.txt"}
	expander := &WordlistExpander{Wordlist: &wl}
	expander.ProcessWordlist()
	if len(*expander.Wordlist) != len(expected) {
		t.Fatalf("Length of wordlist not expected: %d vs %d", len(*expander.Wordlist), len(expected))
	}
	for i, e := range expected {
		if (*expander.Wordlist)[i] != e {
			t.Errorf("Wordlist element mismatch: %s %s", e, (*expander.Wordlist)[i])
		}
	}
}

func TestExpand(t *testing.T) {
	wl := []string{"a", "b"}
	expander := &WordlistExpander{Wordlist: &wl, Adder: func(_ int) {}}
	ch := make(chan *task.Task, 5)
	paths := []string{"/foo", "/bar/"}
	expected := []string{"/foo", "/foo/a", "/foo/b", "/bar/", "/bar/a", "/bar/b"}
	for _, p := range paths {
		ch <- &task.Task{URL: &url.URL{Path: p}}
	}
	close(ch)
	res := expander.Expand(ch)
	for _, exp := range expected {
		if item, ok := <-res; ok {
			if exp != item.URL.Path {
				t.Errorf("Expected %s, got %s.", exp, item.URL.Path)
			}
		} else {
			t.Error("Expected an item, got closed channel!")
		}
	}
	if _, ok := <-res; ok {
		t.Errorf("Expected closed channel, read an item!")
	}
}
