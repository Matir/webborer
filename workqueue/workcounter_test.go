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

package workqueue

import (
	"sync"
	"testing"
)

func TestWorkCounterAdd(t *testing.T) {
	wc := WorkCounter{}
	if wc.todo != 0 {
		t.Fatalf("Expecting 0 todo, got %d", wc.todo)
	}
	wc.Add(1)
	if wc.todo != 1 {
		t.Fatalf("Expecting 1 todo, got %d", wc.todo)
	}
}

func TestWorkCounterDone(t *testing.T) {
	wc := WorkCounter{todo: 1}
	wc.L = &sync.Mutex{}
	wc.Done(1)
	if wc.done != 1 {
		t.Fatalf("Expecting 1 done, got %d", wc.done)
	}

	// Best way I can come up with to test a panic!
	var didPanic bool
	func() {
		defer func() {
			if err := recover(); err != nil {
				didPanic = true
			}
		}()
		wc.Done(1)
	}()
	if !didPanic {
		t.Fatalf("Expected a panic, but it did not!")
	}
}
