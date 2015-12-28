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
	"runtime"
)

// This is the main runner for gobuster.
// TODO: separate the actual scanning from all of the setup steps
func main() {
	EnableStackTraces()

	// Load scan settings
	settings, err := GetScanSettings()
	if err != nil {
		Logf(LogFatal, err.Error())
		return
	}
	ResetLog(settings)
	Logf(LogInfo, "Flags: %s", settings)

	// Set number of threads
	Logf(LogDebug, "Setting GOMAXPROCS to %d.", settings.Threads)
	runtime.GOMAXPROCS(settings.Threads)

	// Load wordlist
	var wordlist []string
	if settings.WordlistPath != "" {
		wordlist, err = ReadWordlistFile(settings.WordlistPath)
	} else {
		wordlist, err = LoadDefaultWordlist()
	}
	if err != nil {
		Logf(LogFatal, "Unable to load wordlist: %s", err.Error())
		return
	}

	// Build an HTTP Client Factory
	Logf(LogDebug, "Creating Client Factory...")
	clientFactory := NewProxyClientFactory(settings.Proxies, settings.Timeout)

	// Starting point
	scope, err := url.Parse(settings.BaseURL)
	if err != nil {
		Logf(LogFatal, "Unable to parse BaseURL: %s", err.Error())
		return
	}
	if scope.Path == "" {
		scope.Path = "/"
	}
	Logf(LogDebug, "BaseURL: %s", scope.String())

	// Setup the main workqueue
	Logf(LogDebug, "Starting work queue...")
	queue := NewWorkQueue(settings.QueueSize, MakeScopeFunc(scope))
	queue.RunInBackground()

	Logf(LogDebug, "Creating expander and filter...")
	expander := Expander{Wordlist: &wordlist, Adder: queue.GetAddCount()}
	expander.ProcessWordlist()
	filter := NewWorkFilter(settings, queue.GetDoneFunc())
	work := filter.Filter(expander.Expand(queue.GetWorkChan()))

	Logf(LogDebug, "Creating results manager...")
	results := make(chan Result, settings.QueueSize)
	resultsManager, err := GetResultsManager(settings)
	if err != nil {
		Logf(LogFatal, "Unable to start results manager: %s", err.Error())
		return
	}

	Logf(LogDebug, "Starting %d workers...", settings.Workers)
	workers := StartWorkers(settings, clientFactory, work, queue.GetAddFunc(), queue.GetDoneFunc(), results)
	if settings.ParseHTML {
		htmlWorker := NewHTMLWorker(queue.GetAddFunc())
		SetPageWorkers(workers, htmlWorker)
	}

	Logf(LogDebug, "Starting results manager...")
	resultsManager.Run(results)

	// Kick things off with the seed URL
	Logf(LogDebug, "Adding starting URL: %s", scope)
	queue.AddURLs(scope)

	// Wait for work to be done
	Logf(LogDebug, "Main goroutine waiting for work...")
	queue.WaitPipe()
	Logf(LogDebug, "Work done.")

	// Cleanup
	queue.InputFinished()
	close(results)

	resultsManager.Wait()
	Logf(LogDebug, "Done!")
}

// Build a function to check if the target URL is in scope.
func MakeScopeFunc(scope *url.URL) func(*url.URL) bool {
	return func(target *url.URL) bool {
		return URLIsSubpath(scope, target)
	}
}
