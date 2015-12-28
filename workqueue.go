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
	"net/url"
	"sync"
)

// WorkQueue is a singleton that maintains the queue of work to be done.
// It reads from one input channel, verifies that the URL is in scope,
// queues it, then writes it to the work channel to be done.
type WorkQueue struct {
	// Elements to be worked on
	head *queueNode
	// End for cheap appends
	tail *queueNode
	// Number of items in queue, for stats
	queueLen int
	// Channel for URLs to be considered
	src chan *url.URL
	// Channel for URLs to be worked on
	dst chan *url.URL
	// filter to determine if a URL should be processed
	filter func(*url.URL) bool
	// channel to track done
	started chan bool
	// counter of work being done
	ctr WorkCounter
}

type queueNode struct {
	// next ptr
	next *queueNode
	// data
	data *url.URL
}

type QueueAddFunc func(...*url.URL)
type QueueAddCount func(int)
type QueueDoneFunc func(int)

func NewWorkQueue(queueSize int, filter func(*url.URL) bool) *WorkQueue {
	q := &WorkQueue{
		src:     make(chan *url.URL, queueSize),
		dst:     make(chan *url.URL, queueSize),
		filter:  filter,
		started: make(chan bool, 1),
	}
	q.ctr.L = &sync.Mutex{}
	return q
}

func (q *WorkQueue) AddURLs(urls ...*url.URL) {
	q.ctr.Add(int64(len(urls)))
	for _, u := range urls {
		q.src <- u
	}
}

func (q *WorkQueue) InputFinished() {
	close(q.src)
}

func (q *WorkQueue) GetWorkChan() <-chan *url.URL {
	return q.dst
}

func (q *WorkQueue) Run() {
	defer close(q.dst)

	q.started <- true
	for true {
		if q.head != nil {
			// If we have work to send, non-blocking read
			select {
			case u, ok := <-q.src:
				if !ok {
					for q.head != nil {
						q.dst <- q.pop()
					}
					return
				}
				if q.filter(u) {
					q.push(u)
				} else {
					q.reject(u)
				}
			case q.dst <- q.peek():
				q.pop()
			}
		} else {
			// Blocking read and non-blocking send
			u, ok := <-q.src
			if !ok {
				return
			}
			if !q.filter(u) {
				q.reject(u)
				continue
			}
			select {
			case q.dst <- u:
			default:
				q.push(u)
			}
		}
	}
}

func (q *WorkQueue) RunInBackground() {
	go q.Run()
}

func (q *WorkQueue) WaitPipe() {
	<-q.started
	q.ctr.L.Lock()
	if q.ctr.todo == q.ctr.done {
		q.ctr.L.Unlock()
		return
	}
	q.ctr.Wait()
}

func (q *WorkQueue) GetAddFunc() QueueAddFunc {
	return func(urls ...*url.URL) {
		q.AddURLs(urls...)
	}
}

func (q *WorkQueue) GetAddCount() QueueAddCount {
	return func(c int) {
		q.ctr.Add(int64(c))
	}
}

func (q *WorkQueue) GetDoneFunc() QueueDoneFunc {
	return func(c int) {
		q.ctr.Done(int64(c))
	}
}

func (q *WorkQueue) reject(u *url.URL) {
	Logf(LogDebug, "Workqueue rejecting %s", u.String())
	q.ctr.Done(1)
}

// Append URL to end of queue
func (q *WorkQueue) push(u *url.URL) {
	node := &queueNode{data: u}
	if q.tail != nil {
		q.tail.next = node
	} else {
		q.head = node
	}
	q.tail = node
	q.queueLen++
}

// Get URL from front of queue
func (q *WorkQueue) pop() *url.URL {
	node := q.head
	if node == nil {
		return nil
	}
	q.head = q.head.next
	if q.head == nil {
		q.tail = nil
	}
	q.queueLen--
	return node.data
}

// Get URL from front of queue without removal
func (q *WorkQueue) peek() *url.URL {
	if q.head != nil {
		return q.head.data
	}
	return nil
}

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
		Logf(LogInfo, "Work counter thinks we're done.")
		ctr.L.Lock()
		defer ctr.L.Unlock()
		ctr.Broadcast()
	}
}

func (ctr *WorkCounter) Stats() {
	Logf(LogDebug, "WorkCounter: %d/%d", ctr.done, ctr.todo)
}
