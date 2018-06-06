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

package filter

import (
	"github.com/Matir/webborer/task"
	"github.com/Matir/webborer/util"
	"github.com/Matir/webborer/workqueue"
	"net/url"
	"strings"
)

// An Expander is responsible for taking input URLs and expanding them to
// include all of the words in the wordlist.
type WordlistExpander struct {
	// List of words to expand
	Wordlist []string
	// Function to count new instances
	adder workqueue.QueueAddCount
}

func NewWordlistExpander(Wordlist []string) *WordlistExpander {
	return &WordlistExpander{
		Wordlist: Wordlist,
	}
}

// Update the wordlist to contain directory & non-directory entries
func (e *WordlistExpander) ProcessWordlist() {
	newList := make([]string, 0)
	for _, w := range e.Wordlist {
		newList = append(newList, w)
		if strings.Contains(w, ".") {
			continue
		}
		if w[len(w)-1] == byte('/') {
			continue
		}
		newList = append(newList, w+"/")
	}
	e.Wordlist = newList
}

func (e *WordlistExpander) Expand(in <-chan *task.Task) <-chan *task.Task {
	out := make(chan *task.Task, cap(in))
	go func() {
		for it := range in {
			out <- it
			e.adder(len(e.Wordlist))
			for _, word := range e.Wordlist {
				t := it.Copy()
				t.URL = ExtendURL(t.URL, word)
				out <- t
			}
		}
		close(out)
	}()

	return out
}

func (e *WordlistExpander) SetAddCount(adder workqueue.QueueAddCount) {
	e.adder = adder
}

func ExtendURL(u *url.URL, tail string) *url.URL {
	extended := *u
	if !util.URLIsDir(u) {
		extended.Path += "/" + tail
	} else {
		extended.Path += tail
	}
	return &extended
}
