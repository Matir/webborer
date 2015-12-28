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
	"net/url"
	"testing"
)

func TestWorkqueueBasic(t *testing.T) {
	filter := func(_ *url.URL) bool { return true }

	queue := NewWorkQueue(5, filter)
	queue.RunInBackground()
	fmt.Println("Adding tasks...")
	for i := 0; i < 20; i++ {
		s := fmt.Sprintf("%d", i)
		u := &url.URL{Path: s}
		queue.AddURLs(u)
	}
	queue.InputFinished()
	fmt.Println("Getting tasks...")
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
	fmt.Println("Waiting...")
	queue.WaitPipe()
}
