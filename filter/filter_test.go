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
	"github.com/Matir/gobuster/client/mock"
	"github.com/Matir/gobuster/settings"
	"net/url"
	"testing"
)

func TestFilterDuplicates(t *testing.T) {
	src := make(chan *url.URL, 5)
	for _, p := range []string{"/a", "/b", "/a", "/c", "/a"} {
		src <- &url.URL{Path: p}
	}
	dupes := 0
	dupefunc := func(i int) { dupes += i }
	filter := NewWorkFilter(&settings.ScanSettings{}, dupefunc)
	close(src)
	out := filter.RunFilter(src)
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

func TestFilterExclusion(t *testing.T) {
	src := make(chan *url.URL, 5)
	src <- &url.URL{Path: "/a"}
	src <- &url.URL{Path: "/b"}
	dupefunc := func(_ int) {}
	ss := &settings.ScanSettings{
		ExcludePaths: []string{
			"/a",
		},
	}
	filter := NewWorkFilter(ss, dupefunc)
	close(src)
	out := filter.RunFilter(src)
	if u, ok := <-out; ok {
		if u.Path != "/b" {
			t.Errorf("Expected /b, got %v", u)
		}
	} else {
		t.Errorf("Expected output, got closed channel.")
	}
	if u, ok := <-out; ok {
		t.Errorf("Expected closed channel, got %v instead.", u)
	}
}

func TestFilterParseFail(t *testing.T) {
	ss := &settings.ScanSettings{
		ExcludePaths: []string{
			"://",
		},
	}
	wf := NewWorkFilter(ss, func(_ int) {})
	if len(wf.exclusions) != 0 {
		t.Error("Expected error parsing exclusion, but got none.")
	}
}

func TestRobotsFilter_Success(t *testing.T) {
	wf := NewWorkFilter(&settings.ScanSettings{}, func(_ int) {})
	client := &mock.MockClient{NextResponse: mock.MockRobotsResponse()}
	cf := &mock.MockClientFactory{NextClient: client}
	u, _ := url.Parse("http://localhost/")
	wf.AddRobotsFilter([]*url.URL{u}, cf)
	if len(wf.exclusions) != 1 {
		t.Errorf("Expected one exclusion, got %d", len(wf.exclusions))
	}
}

func TestRobotsFilter_Fail(t *testing.T) {
	wf := NewWorkFilter(&settings.ScanSettings{}, func(_ int) {})
	cf := &mock.MockClientFactory{}
	u, _ := url.Parse("http://localhost/")
	wf.AddRobotsFilter([]*url.URL{u}, cf)
	if len(wf.exclusions) != 0 {
		t.Errorf("Expected no exclusions, got %d", len(wf.exclusions))
	}
}
