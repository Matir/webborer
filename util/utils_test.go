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
	"net/url"
	"testing"
)

func BenchmarkByte(b *testing.B) {
	u := "abcdef/"
	slash := byte('/')
	for i := 0; i < b.N; i++ {
		_ = u[len(u)-1] == slash
	}
}

func BenchmarkString(b *testing.B) {
	u := "abcdef/"
	for i := 0; i < b.N; i++ {
		_ = u[len(u)-1:] == "/"
	}
}

func TestURLIsDir(t *testing.T) {
	u := &url.URL{Path: "/abcdef/"}
	if !URLIsDir(u) {
		t.Errorf("%s should be a directory.", u.Path)
	}
	u.Path = "/abcdef"
	if URLIsDir(u) {
		t.Errorf("%s shouldn't be a directory.", u.Path)
	}
	u = &url.URL{Host: "localhost"} // no path
	if !URLIsDir(u) {
		t.Errorf("%s should be a directory.", u.String())
	}
}

func TestStatusCodeGroup(t *testing.T) {
	tests := map[int]int{
		200: 200,
		201: 200,
		299: 200,
		300: 300,
		305: 300,
		404: 400,
	}
	for k, v := range tests {
		if StatusCodeGroup(k) != v {
			t.Errorf("%d should be in group %d.", k, v)
		}
	}
}

func TestDedupeStrings(t *testing.T) {
	testdata := []string{
		"a",
		"b",
		"c",
		"a",
		"a",
		"c",
		"d",
	}
	testresult := []string{
		"a",
		"b",
		"c",
		"d",
	}
	result := DedupeStrings(testdata)
	if !slicesEqual(testresult, result) {
		t.Errorf("Slices not equal: %v, %v", result, testresult)
	}
}

func slicesEqual(a, b []string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestURLIsSubpath(t *testing.T) {
	parentAbs, _ := url.Parse("http://localhost/foo/bar")
	parentRel, _ := url.Parse("/foo/bar")
	parentRoot, _ := url.Parse("http://localhost/")
	wildcard, _ := url.Parse("/")
	parents := []*url.URL{parentAbs, parentRel, parentRoot, wildcard}
	tests := map[string][]bool{
		"http://localhost/foo/bar":         {true, true, true, true},
		"http://localhost/foo/bar/":        {true, true, true, true},
		"http://localhost/foo/bar/baz":     {true, true, true, true},
		"http://localhost/foo":             {false, false, true, true},
		"http://localhost/foo/barn":        {false, false, true, true},
		"file://localhost/foo/bar":         {false, true, false, true},
		"https://localhost/foo/bar":        {false, true, false, true},
		"http://127.0.0.1/foo/bar":         {false, true, false, true},
		"http://localhost/foo/bar/..":      {false, false, true, true},
		"http://localhost/foo/baz/../bar/": {true, true, true, true},
	}
	for i, parent := range parents {
		for child, expects := range tests {
			curl, _ := url.Parse(child)
			value := URLIsSubpath(parent, curl)
			if value != expects[i] {
				var expString string
				if expects[i] {
					expString = "should"
				} else {
					expString = "shouldn't"
				}
				t.Errorf("Parent: %s, child %s, %s match subpath.",
					parent.String(), curl.String(), expString)
			}
		}
	}
}

func BenchmarkURLIsSubpath(b *testing.B) {
	parent, _ := url.Parse("http://localhost/foo/bar")
	child, _ := url.Parse("http://localhost/foo/bar/baz")
	child2, _ := url.Parse("http://localhost/bang/baz")
	for i := 0; i < b.N/2; i++ {
		URLIsSubpath(parent, child)
		URLIsSubpath(parent, child2)
	}
}

func TestGetParentsPathString(t *testing.T) {
	pathA := "/abc/def/ghi/jkl.txt"
	expectedA := []string{"/abc", "/abc/def", "/abc/def/ghi"}
	resultsA := getParentPathsString(pathA)
	if !slicesEqual(expectedA, resultsA) {
		t.Errorf("%v != %v", expectedA, resultsA)
	}

	pathB := "/abc/def/ghi/jkl/"
	expectedB := []string{"/abc", "/abc/def", "/abc/def/ghi"}
	resultsB := getParentPathsString(pathB)
	if !slicesEqual(expectedB, resultsB) {
		t.Errorf("%v != %v", expectedB, resultsB)
	}
}
