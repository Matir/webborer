package main

import (
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
)

var slash = byte('/')

func URLIsDir(u *url.URL) bool {
	l := len(u.Path)
	if l == 0 {
		return true
	}
	return u.Path[l-1] == slash
}

// Returns true if this path should be spidered more
func KeepSpidering(code int) bool {
	return (StatusCodeGroup(code) == 200 ||
		code == http.StatusUnauthorized ||
		code == http.StatusForbidden)
}

// Returns true if this is a "useful" result
func FoundSomething(code int) bool {
	return (code != http.StatusNotFound &&
		code != http.StatusGone &&
		code != http.StatusBadGateway &&
		code != http.StatusServiceUnavailable &&
		code != http.StatusGatewayTimeout)
}

// Find the group (200, 300, 400, 500, ...) this status code belongs to
func StatusCodeGroup(code int) int {
	return (code / 100) * 100
}

// Enable stack traces on SIGQUIT
// Thanks to:
func EnableStackTraces() {
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGQUIT)
		buf := make([]byte, 1<<20)
		for {
			<-sigs
			runtime.Stack(buf, true)
			Logf(LogDebug, "=== received SIGQUIT ===\n*** goroutine dump...\n%s\n*** end\n", buf)
		}
	}()
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
	Logf(LogDebug, "Subpath check: Parent: %s, child %s.", parent.String(), child.String())
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
		Logf(LogDebug, "Reject for differing paths: %s, %s", cPath, pPath)
		return false
	}
	return cPath[len(pPath)] == slash
}
