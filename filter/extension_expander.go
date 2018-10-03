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
	"fmt"
	"github.com/matir/webborer/task"
	"github.com/matir/webborer/workqueue"
	"net/url"
	"strings"
)

// Expand extensions when none are given
type ExtensionExpander struct {
	extensions []string
	adder      workqueue.QueueAddCount
}

func NewExtensionExpander(extensions []string) *ExtensionExpander {
	return &ExtensionExpander{extensions: extensions}
}

func (e *ExtensionExpander) SetAddCount(adder workqueue.QueueAddCount) {
	e.adder = adder
}

func (e *ExtensionExpander) Expand(in <-chan *task.Task) <-chan *task.Task {
	outChan := make(chan *task.Task)
	go func() {
		defer close(outChan)
		numExtensions := len(e.extensions)
		for it := range in {
			// Un modified form
			outChan <- it
			if hasExtension(it.URL) {
				continue
			}
			if isDirectory(it.URL) {
				continue
			}
			e.adder(numExtensions)
			for _, ext := range e.extensions {
				t := it.Copy()
				t.URL.Path = fmt.Sprintf("%s.%s", it.URL.Path, ext)
				outChan <- t
			}
		}
	}()
	return outChan
}

func hasExtension(URL *url.URL) bool {
	if slashPos := strings.LastIndex(URL.Path, "/"); slashPos > -1 {
		return strings.LastIndex(URL.Path, ".") > slashPos
	}
	return false
}

func isDirectory(URL *url.URL) bool {
	return strings.HasSuffix(URL.Path, "/")
}
