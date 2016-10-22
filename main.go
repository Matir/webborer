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
	words, err = wordlist.LoadWordlist(settings.WordlistPath)
	if err != nil {
		logging.Logf(logging.LogFatal, "Unable to load wordlist: %s", err.Error())
		return
	}

	// Build an HTTP Client Factory
	logging.Logf(logging.LogDebug, "Creating Client Factory...")
	clientFactory := client.NewProxyClientFactory(settings.Proxies, settings.Timeout, settings.UserAgent)

	// Starting point
	scope, err := settings.GetScopes()
	if err != nil {
		logging.Logf(logging.LogFatal, err.Error())
		return
	}

	// Setup the main workqueue
	logging.Logf(logging.LogDebug, "Starting work queue...")
	queue := workqueue.NewWorkQueue(settings.QueueSize, MakeScopeFunc(scope, settings.AllowHTTPSUpgrade))
	queue.RunInBackground()

	logging.Logf(logging.LogDebug, "Creating expander and filter...")
	expander := filter.Expander{Wordlist: &words, Adder: queue.GetAddCount()}
	expander.ProcessWordlist()
	filter := filter.NewWorkFilter(settings, queue.GetDoneFunc())

	// Check robots mode
	if settings.RobotsMode == ss.ObeyRobots {
		filter.AddRobotsFilter(scope, clientFactory)
	}

	work := filter.RunFilter(expander.Expand(queue.GetWorkChan()))

	logging.Logf(logging.LogDebug, "Creating results manager...")
	rchan := make(chan results.Result, settings.QueueSize)
	resultsManager, err := results.GetResultsManager(settings)
	if err != nil {
		logging.Logf(logging.LogFatal, "Unable to start results manager: %s", err.Error())
		return
	}

	logging.Logf(logging.LogDebug, "Starting %d workers...", settings.Workers)
	worker.StartWorkers(settings, clientFactory, work, queue.GetAddFunc(), queue.GetDoneFunc(), rchan)

	logging.Logf(logging.LogDebug, "Starting results manager...")
	resultsManager.Run(rchan)

	// Kick things off with the seed URL
	logging.Logf(logging.LogDebug, "Adding starting URLs: %v", scope)
	queue.AddURLs(scope...)

	// Potentially seed from robots
	if settings.RobotsMode == ss.SeedRobots {
		queue.SeedFromRobots(scope, clientFactory)
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
func MakeScopeFunc(scope []*url.URL, allowUpgrades bool) func(*url.URL) bool {
	allowedScopes := make([]*url.URL, len(scope))
	copy(allowedScopes, scope)
	if allowUpgrades {
		for _, scopeURL := range scope {
			if scopeURL.Scheme == "http" {
				deref := *scopeURL
				clone := &deref // Can't find a way to do this in one statement
				clone.Scheme = "https"
				allowedScopes = append(allowedScopes, clone)
			}
		}
	}
	return func(target *url.URL) bool {
		for _, scopeURL := range scope {
			if util.URLIsSubpath(scopeURL, target) {
				return true
			}
		}
		return false
	}
}
