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

package filter

import (
	"fmt"
	"github.com/Matir/gobuster/settings"
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
	filter := NewWorkFilter(&settings.ScanSettings{}, dupefunc)
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
