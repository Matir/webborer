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
	"github.com/Matir/gobuster/client"
	"github.com/Matir/gobuster/filter"
	"github.com/Matir/gobuster/logging"
	"github.com/Matir/gobuster/results"
	"github.com/Matir/gobuster/robots"
	ss "github.com/Matir/gobuster/settings"
	"github.com/Matir/gobuster/util"
	"github.com/Matir/gobuster/wordlist"
	"github.com/Matir/gobuster/worker"
	"github.com/Matir/gobuster/workqueue"
	"net/url"
	"runtime"
)

// This is the main runner for gobuster.
// TODO: separate the actual scanning from all of the setup steps
func main() {
	util.EnableStackTraces()

	// Load scan settings
	settings, err := ss.GetScanSettings()
	if err != nil {
		logging.Logf(logging.LogFatal, err.Error())
		return
	}
	logging.ResetLog(settings.LogfilePath, settings.LogLevel)
	logging.Logf(logging.LogInfo, "Flags: %s", settings)

	// Set number of threads
	logging.Logf(logging.LogDebug, "Setting GOMAXPROCS to %d.", settings.Threads)
	runtime.GOMAXPROCS(settings.Threads)

	// Load wordlist
	var words []string
	if settings.WordlistPath != "" {
		words, err = wordlist.ReadWordlistFile(settings.WordlistPath)
	} else {
		words, err = wordlist.LoadDefaultWordlist()
	}
	if err != nil {
		logging.Logf(logging.LogFatal, "Unable to load wordlist: %s", err.Error())
		return
	}

	// Build an HTTP Client Factory
	logging.Logf(logging.LogDebug, "Creating Client Factory...")
	clientFactory := client.NewProxyClientFactory(settings.Proxies, settings.Timeout, settings.UserAgent)

	// Starting point
	scope := make([]*url.URL, len(settings.BaseURLs))
	for i, baseURL := range settings.BaseURLs {
		scope[i], err = url.Parse(baseURL)
		if err != nil {
			logging.Logf(logging.LogFatal, "Unable to parse BaseURL (%s): %s", baseURL, err.Error())
			return
		}
		if scope[i].Path == "" {
			scope[i].Path = "/"
		}
		logging.Logf(logging.LogDebug, "Added BaseURL: %s", scope[i].String())
	}

	// Setup the main workqueue
	logging.Logf(logging.LogDebug, "Starting work queue...")
	queue := workqueue.NewWorkQueue(settings.QueueSize, MakeScopeFunc(scope))
	queue.RunInBackground()

	logging.Logf(logging.LogDebug, "Creating expander and filter...")
	expander := filter.Expander{Wordlist: &words, Adder: queue.GetAddCount()}
	expander.ProcessWordlist()
	filter := filter.NewWorkFilter(settings, queue.GetDoneFunc())
	work := filter.Filter(expander.Expand(queue.GetWorkChan()))

	// Check robots mode
	if settings.RobotsMode == ss.ObeyRobots {
		for _, scopeURL := range scope {
			robotsData, err := robots.GetRobotsForURL(scopeURL, clientFactory)
			if err != nil {
				logging.Logf(logging.LogWarning, "Unable to get robots.txt data: %s", err)
			} else {
				for _, disallowed := range robotsData.GetForUserAgent(settings.UserAgent) {
					if pathURL, err := url.Parse(disallowed); err != nil {
						disallowedURL := scopeURL.ResolveReference(pathURL)
						filter.FilterURL(disallowedURL)
					}
				}
			}
		}
	}

	logging.Logf(logging.LogDebug, "Creating results manager...")
	rchan := make(chan results.Result, settings.QueueSize)
	resultsManager, err := results.GetResultsManager(settings)
	if err != nil {
		logging.Logf(logging.LogFatal, "Unable to start results manager: %s", err.Error())
		return
	}

	logging.Logf(logging.LogDebug, "Starting %d workers...", settings.Workers)
	workers := worker.StartWorkers(settings, clientFactory, work, queue.GetAddFunc(), queue.GetDoneFunc(), rchan)
	if settings.ParseHTML {
		htmlWorker := worker.NewHTMLWorker(queue.GetAddFunc())
		worker.SetPageWorkers(workers, htmlWorker)
	}

	logging.Logf(logging.LogDebug, "Starting results manager...")
	resultsManager.Run(rchan)

	// Kick things off with the seed URL
	for _, scopeURL := range scope {
		logging.Logf(logging.LogDebug, "Adding starting URL: %s", scopeURL)
		queue.AddURLs(scopeURL)
	}

	// Potentially seed from robots
	if settings.RobotsMode == ss.SeedRobots {
		for _, scopeURL := range scope {
			robotsData, err := robots.GetRobotsForURL(scopeURL, clientFactory)
			if err != nil {
				logging.Logf(logging.LogWarning, "Unable to get robots.txt data: %s", err)
			} else {
				for _, path := range robotsData.GetAllPaths() {
					if pathURL, err := url.Parse(path); err != nil {
						// Filter will handle if this is out of scope
						queue.AddURLs(scopeURL.ResolveReference(pathURL))
					}
				}
			}
		}
	}

	// Wait for work to be done
	logging.Logf(logging.LogDebug, "Main goroutine waiting for work...")
	queue.WaitPipe()
	logging.Logf(logging.LogDebug, "Work done.")

	// Cleanup
	queue.InputFinished()
	close(rchan)

	resultsManager.Wait()
	logging.Logf(logging.LogDebug, "Done!")
}

// Build a function to check if the target URL is in scope.
func MakeScopeFunc(scope []*url.URL) func(*url.URL) bool {
	return func(target *url.URL) bool {
		for _, scopeURL := range scope {
			if util.URLIsSubpath(scopeURL, target) {
				return true
			}
		}
		return false
	}
}
