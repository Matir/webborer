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

// WebBorer is a directory-enumeration tool written in Go.
package main

import (
	"github.com/Matir/webborer/client"
	"github.com/Matir/webborer/filter"
	"github.com/Matir/webborer/logging"
	"github.com/Matir/webborer/results"
	ss "github.com/Matir/webborer/settings"
	"github.com/Matir/webborer/task"
	"github.com/Matir/webborer/util"
	"github.com/Matir/webborer/wordlist"
	"github.com/Matir/webborer/worker"
	"github.com/Matir/webborer/workqueue"
	"runtime"
)

// Load settings from flags
func loadSettings() (*ss.ScanSettings, error) {
	// Load scan settings
	settings, err := ss.GetScanSettings()
	if err != nil {
		logging.Logf(logging.LogFatal, err.Error())
		return nil, err
	}
	logging.ResetLog(settings.LogfilePath, settings.LogLevel)
	logging.Logf(logging.LogInfo, "Flags: %s", settings)
	return settings, nil
}

// This is the main runner for webborer.
// TODO: separate the actual scanning from all of the setup steps
func main() {
	util.EnableStackTraces()

	settings, err := loadSettings()
	if err != nil {
		return
	}

	// Enable CPU profiling
	var cpuProfStop func()
	if settings.DebugCPUProf {
		cpuProfStop = util.EnableCPUProfiling()
	}

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
	clientFactory, err := client.NewProxyClientFactory(settings.Proxies, settings.Timeout, settings.UserAgent)
	if err != nil {
		logging.Logf(logging.LogFatal, "Unable to build client factory: %s", err.Error())
		return
	}
	clientFactory.SetUsernamePassword(settings.HTTPUsername, settings.HTTPPassword)

	// Starting point
	scope, err := settings.GetScopes()
	if err != nil {
		logging.Logf(logging.LogFatal, err.Error())
		return
	}

	// Setup the main workqueue
	logging.Logf(logging.LogDebug, "Starting work queue...")
	queue := workqueue.NewWorkQueue(settings.QueueSize, scope, settings.AllowHTTPSUpgrade)
	queue.RunInBackground()

	logging.Logf(logging.LogDebug, "Creating expander and filter...")
	var expander filter.Expander
	switch settings.RunMode {
	case ss.RunModeEnumeration:
		wlexpander := filter.NewWordlistExpander(words)
		wlexpander.ProcessWordlist()
		expander = wlexpander
	case ss.RunModeDotProduct:
		dpexpander := filter.NewDotProductExpander(words)
		expander = dpexpander
	default:
		panic("Unknown run mode!")
	}
	expander.SetAddCount(queue.GetAddCount())

	headerExpander := filter.NewHeaderExpander(settings.OptionalHeader.Header())
	headerExpander.SetAddCount(queue.GetAddCount())
	extensionExpander := filter.NewExtensionExpander(settings.Extensions)
	extensionExpander.SetAddCount(queue.GetAddCount())

	filter := filter.NewWorkFilter(settings, queue.GetDoneFunc())

	// Check robots mode
	if settings.RobotsMode == ss.ObeyRobots {
		filter.AddRobotsFilter(scope, clientFactory)
	}

	// filter paths after expansion
	logging.Debugf("Starting expansion and filtering...")
	workChan := queue.GetWorkChan()
	workChan = expander.Expand(workChan)
	workChan = headerExpander.Expand(workChan)
	workChan = extensionExpander.Expand(workChan)
	workChan = filter.RunFilter(workChan)

	logging.Logf(logging.LogDebug, "Creating results manager...")
	rchan := make(chan *results.Result, settings.QueueSize)
	resultsManager, err := results.GetResultsManager(settings)
	if err != nil {
		logging.Logf(logging.LogFatal, "Unable to start results manager: %s", err.Error())
		return
	}

	logging.Logf(logging.LogDebug, "Starting %d workers...", settings.Workers)
	worker.StartWorkers(settings, clientFactory, workChan, queue.GetAddFunc(), queue.GetDoneFunc(), rchan)

	logging.Logf(logging.LogDebug, "Starting results manager...")
	resultsManager.Run(rchan)

	// Kick things off with the seed URL
	logging.Logf(logging.LogDebug, "Adding starting URLs: %v", scope)
	task.SetDefaultHeader(settings.Header.Header())
	tasks := make([]*task.Task, 0, len(scope))
	for _, s := range scope {
		tasks = append(tasks, task.NewTaskFromURL(s))
	}
	queue.AddTasks(tasks...)

	// Add a progress bar?
	if settings.ProgressBar {
		initProgressBar(queue.GetCounter())
	}

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

	logging.Debugf("Waiting for results manager.")
	resultsManager.Wait()
	if cpuProfStop != nil {
		cpuProfStop()
	}
	logging.Logf(logging.LogDebug, "Done!")
}
