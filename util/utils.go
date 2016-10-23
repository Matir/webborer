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

package util

import (
	"github.com/Matir/gobuster/logging"
	"net/url"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
)

var slash = byte('/')
var dot = byte('.')

func URLIsDir(u *url.URL) bool {
	l := len(u.Path)
	if l == 0 {
		return true
	}
	return u.Path[l-1] == slash
}

func URLHasExtension(u *url.URL) bool {
	return strings.LastIndexByte(u.Path, dot) > strings.LastIndexByte(u.Path, slash)
}

// Find the group (200, 300, 400, 500, ...) this status code belongs to
func StatusCodeGroup(code int) int {
	return (code / 100) * 100
}

// Enable stack traces on SIGQUIT
// Returns a function that can be used to disable stack traces.
func EnableStackTraces() func() {
	sigs := make(chan os.Signal, 1)
	go func() {
		signal.Notify(sigs, syscall.SIGQUIT)
		buf := make([]byte, 1<<20)
		for range sigs {
			runtime.Stack(buf, true)
			logging.Logf(logging.LogDebug, "=== received SIGQUIT ===\n*** goroutine dump...\n%s\n*** end\n", buf)
		}
	}()
	return func() {
		signal.Stop(sigs)
		close(sigs)
	}
}

// Deduplicate a slice of strings
func DedupeStrings(s []string) []string {
	table := make(map[string]bool)
	out := make([]string, 0)
	for _, v := range s {
		if _, ok := table[v]; !ok {
			out = append(out, v)
			table[v] = true
		}
	}
	return out
}

// Determine if one path is a subpath of another path
// Only considers the host and scheme if they are non-empty in the parent
// Identical paths are considered subpaths of each other
func URLIsSubpath(parent, child *url.URL) bool {
	logging.Logf(logging.LogDebug, "Subpath check: Parent: %s, child %s.", parent.String(), child.String())
	if parent.Scheme != "" && child.Scheme != parent.Scheme {
		return false
	}
	if parent.Host != "" && child.Host != parent.Host {
		return false
	}
	if parent.Path == "/" {
		// Everything is in this path
		return true
	}
	// Now split the path
	pPath := path.Clean(parent.Path)
	cPath := path.Clean(child.Path)
	if len(cPath) < len(pPath) {
		return false
	}
	if cPath == pPath {
		return true
	}
	if !strings.HasPrefix(cPath, pPath) {
		logging.Logf(logging.LogDebug, "Reject for differing paths: %s, %s", cPath, pPath)
		return false
	}
	return cPath[len(pPath)] == slash
}

// Get the parent paths of a given path
func GetParentPaths(child *url.URL) []*url.URL {
	childPath := strings.TrimRight(child.Path, "/")
	var results []*url.URL
	for _, path := range getParentPathsString(childPath) {
		parentURL := *child
		parentURL.Path = path
		results = append(results, &parentURL)
	}
	return results
}

func getParentPathsString(childPath string) []string {
	splitPath := strings.Split(strings.TrimRight(childPath, "/"), "/")
	var results []string
	for i := 2; i < len(splitPath); i++ {
		results = append(results, strings.Join(splitPath[:i], "/"))
	}
	return results
}
