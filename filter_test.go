package main

import (
	"fmt"
	"net/url"
	"testing"
)

func TestFilterDuplicates(t *testing.T) {
	fmt.Println("Adding tasks...")
	src := make(chan *url.URL, 5)
	src <- &url.URL{Path: "/a"}
	src <- &url.URL{Path: "/b"}
	src <- &url.URL{Path: "/a"}
	src <- &url.URL{Path: "/c"}
	src <- &url.URL{Path: "/a"}
	dupes := 0
	dupefunc := func(i int) { dupes += i }
	filter := NewWorkFilter(&ScanSettings{}, dupefunc)
	fmt.Println("Starting filtering...")
	close(src)
	out := filter.Filter(src)
	for _, p := range []string{"/a", "/b", "/c"} {
		if u, ok := <-out; ok {
			if u.Path != p {
				t.Errorf("Expected %s, got %s.", p, u.Path)
			}
		} else {
			t.Error("Expected output, channel was closed.")
		}
	}
	if _, ok := <-out; ok {
		t.Error("Expected closed channel, got read.")
	}
	if dupes != 2 {
		t.Errorf("Expected 2 dupes, got %d", dupes)
	}
}
