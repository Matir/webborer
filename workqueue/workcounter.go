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
	"github.com/matir/webborer/logging"
	"sync"
)

// Count work to do and work done
type WorkCounter struct {
	todo   int64
	done   int64
	doneCb func(done, total int64)
	sync.Mutex
	sync.Cond
}

// Increment the count to be done (input)
func (ctr *WorkCounter) Add(todo int64) {
	ctr.Lock()
	defer ctr.Unlock()
	ctr.todo += todo
	ctr.Stats()
}

// Increment the count that is done (output)
func (ctr *WorkCounter) Done(done int64) {
	ctr.Lock()
	defer ctr.Unlock()
	ctr.done += done
	ctr.Stats()
	if ctr.done > ctr.todo {
		panic("Done exceeded todo in WorkCounter!")
	}
	if ctr.done == ctr.todo {
		// Mark done
		logging.Logf(logging.LogInfo, "Work counter thinks we're done.")
		// These are part of the sync.Cond
		ctr.L.Lock()
		defer ctr.L.Unlock()
		ctr.Broadcast()
	}
}

// Update the stats of the counter
func (ctr *WorkCounter) Stats() {
	logging.Logf(logging.LogDebug, "WorkCounter: %d/%d", ctr.done, ctr.todo)
	if ctr.doneCb != nil {
		ctr.doneCb(ctr.done, ctr.todo)
	}
}

// Set the status callback for this workcounter
func (ctr *WorkCounter) SetStatusCallback(f func(int64, int64)) {
	ctr.doneCb = f
}
