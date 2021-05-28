// Copyright 2018 Google Inc. All Rights Reserved.
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
	"github.com/Matir/webborer/workqueue"
	"net/http"
)

// Header expander tries each of the headers known to it in turn
type HeaderExpander struct {
	Header http.Header
	adder  workqueue.QueueAddCount
}

func NewHeaderExpander(header http.Header) *HeaderExpander {
	return &HeaderExpander{
		Header: header,
	}
}

func (e *HeaderExpander) SetAddCount(adder workqueue.QueueAddCount) {
	e.adder = adder
}

func (e *HeaderExpander) Expand(in <-chan *task.Task) <-chan *task.Task {
	outChan := make(chan *task.Task)
	go func() {
		defer close(outChan)
		for it := range in {
			// Un modified form
			outChan <- it
			for k, vals := range e.Header {
				for _, v := range vals {
					newIt := it.Copy()
					newIt.Header.Set(k, v)
					e.adder(1)
					outChan <- newIt
				}
			}
		}
	}()
	return outChan
}
