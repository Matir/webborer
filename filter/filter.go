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

// Package filter provides the ability to filter and modify a channel of URLs.
// This includes filtering exclusions, duplicates, and expanding the channel
// based on the wordlist.
package filter

import (
	"github.com/Matir/gobuster/client"
	"github.com/Matir/gobuster/logging"
	"github.com/Matir/gobuster/robots"
	ss "github.com/Matir/gobuster/settings"
	"github.com/Matir/gobuster/util"
	"github.com/Matir/gobuster/workqueue"
	"net/url"
)

// WorkFilter is responsible for making sure that a given URL is only tested
// once, and also for applying any exclusion rules to prevent URLs from being
// scanned.
type WorkFilter struct {
	done     map[string]bool
	settings *ss.ScanSettings
	// Excluded paths
	exclusions []*url.URL
	// Count the work that has been dropped
	counter workqueue.QueueDoneFunc
}

func NewWorkFilter(settings *ss.ScanSettings, counter workqueue.QueueDoneFunc) *WorkFilter {
	wf := &WorkFilter{done: make(map[string]bool), settings: settings, counter: counter}
	wf.exclusions = make([]*url.URL, 0, len(settings.ExcludePaths))
	for _, path := range settings.ExcludePaths {
		if u, err := url.Parse(path); err != nil {
			logging.Logf(logging.LogError, "Unable to parse exclusion path: %s (%s)", path, err.Error())
		} else {
			wf.FilterURL(u)
		}
	}
	return wf
}

// Apply a filter to a channel of URLs.  Runs asynchronously.
func (f *WorkFilter) RunFilter(src <-chan *url.URL) <-chan *url.URL {
	c := make(chan *url.URL, f.settings.QueueSize)
	go func() {
	taskLoop:
		for task := range src {
			taskURL := task.String()
			if _, ok := f.done[taskURL]; ok {
				f.reject(task, "already done")
				continue
			}
			f.done[taskURL] = true
			for _, exclusion := range f.exclusions {
				if util.URLIsSubpath(exclusion, task) {
					f.reject(task, "excluded")
					continue taskLoop
				}
			}
			c <- task
		}
		close(c)
	}()
	return c
}

// Add another URL to filter
func (f *WorkFilter) FilterURL(u *url.URL) {
	f.exclusions = append(f.exclusions, u)
}

// Filter data from robots.txt
func (f *WorkFilter) AddRobotsFilter(scope []*url.URL, clientFactory client.ClientFactory) {
	for _, scopeURL := range scope {
		logging.Logf(logging.LogDebug, "Getting robots.txt exclusions for %s", scopeURL)
		robotsData, err := robots.GetRobotsForURL(scopeURL, clientFactory)
		if err != nil {
			logging.Logf(logging.LogWarning, "Unable to get robots.txt data: %s", err)
		} else {
			for _, disallowed := range robotsData.GetForUserAgent(f.settings.UserAgent) {
				disallowedURL := *scopeURL
				disallowedURL.Path = disallowed
				logging.Logf(logging.LogDebug, "Disallowing URL by robots: %s", &disallowedURL)
				f.FilterURL(&disallowedURL)
			}
		}
	}
}

// Task that can't be used, but should be counted as terminated.
func (f *WorkFilter) reject(u *url.URL, reason string) {
	logging.Logf(logging.LogDebug, "Filter rejected %s: %s.", u.String(), reason)
	f.counter(1)
}
