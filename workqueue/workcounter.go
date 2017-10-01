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
	"github.com/Matir/webborer/logging"
	"sync"
)

// Count work to do and work done
type WorkCounter struct {
	todo int64
	done int64
	sync.Mutex
	sync.Cond
}

func (ctr *WorkCounter) Add(todo int64) {
	ctr.Lock()
	defer ctr.Unlock()
	ctr.todo += todo
	ctr.Stats()
}

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
		ctr.L.Lock()
		defer ctr.L.Unlock()
		ctr.Broadcast()
	}
}

func (ctr *WorkCounter) Stats() {
	logging.Logf(logging.LogDebug, "WorkCounter: %d/%d", ctr.done, ctr.todo)
}
