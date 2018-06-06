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

package task

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
)

type Task struct {
	URL    *url.URL
	Host   string
	Header http.Header

	// Mutex to protect map & data structures
	sync.Mutex
}

var defaultHeader http.Header

func NewTaskFromURL(src *url.URL) *Task {
	return &Task{
		URL:    src,
		Header: defaultHeader,
	}
}

func (t *Task) String() string {
	base := t.URL.String()
	if t.Host != "" {
		base = fmt.Sprintf("%s (%s)", base, t.Host)
	}
	return base
}

func (t *Task) Copy() *Task {
	t.Lock()
	defer t.Unlock()
	tmpU := *t.URL
	newT := &Task{
		Host: t.Host,
		URL:  &tmpU,
	}
	newT.Header = make(http.Header)
	for k, v := range t.Header {
		newT.Header[k] = v[:] // Need to copy the slice
	}
	return newT
}

func SetDefaultHeader(header http.Header) {
	defaultHeader = header
}
