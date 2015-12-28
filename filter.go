package main

import "net/url"

// WorkFilter is responsible for making sure that a given URL is only tested
// once, and also for applying any exclusion rules to prevent URLs from being
// scanned.
type WorkFilter struct {
	done     map[string]bool
	settings *ScanSettings
	// Excluded paths
	exclusions []*url.URL
	// Count the work that has been dropped
	counter QueueDoneFunc
}

func NewWorkFilter(settings *ScanSettings, counter QueueDoneFunc) *WorkFilter {
	wf := &WorkFilter{done: make(map[string]bool), settings: settings, counter: counter}
	wf.exclusions = make([]*url.URL, 0, len(settings.ExcludePaths))
	for _, path := range settings.ExcludePaths {
		if u, err := url.Parse(path); err != nil {
			Logf(LogError, "Unable to parse exclusion path: %s (%s)", path, err.Error())
		} else {
			wf.exclusions = append(wf.exclusions, u)
		}
	}
	return wf
}

func (f *WorkFilter) Filter(src <-chan *url.URL) <-chan *url.URL {
	c := make(chan *url.URL, f.settings.QueueSize)
	go func() {
	taskLoop:
		for task := range src {
			taskURL := task.String()
			if _, ok := f.done[taskURL]; ok {
				f.reject(task)
				continue
			}
			f.done[taskURL] = true
			for _, exclusion := range f.exclusions {
				if URLIsSubpath(exclusion, task) {
					f.reject(task)
					continue taskLoop
				}
			}
			c <- task
		}
		close(c)
	}()
	return c
}

// Task that can't be used
func (f *WorkFilter) reject(u *url.URL) {
	Logf(LogDebug, "Filter rejected %s.", u.String())
	f.counter(1)
}
