// Copyright 2016 Google Inc. All Rights Reserved.
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

package robots

import (
	"github.com/Matir/gobuster/client/mock"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
)

func loadTestRobots(t *testing.T) *RobotsData {
	testFile, err := os.Open("testdata/test_robots.txt")
	if err != nil {
		t.Fatalf("Could not open testdata: %s", err)
	}
	defer testFile.Close()
	buf, err := ioutil.ReadAll(testFile)
	if err != nil {
		t.Fatalf("Could not read testdata: %s", err)
	}
	parsed, err := ParseRobotsTxt(buf)
	if err != nil {
		t.Fatalf("Could not parse test data: %s", err)
	}
	return parsed
}

func TestParseRobots(t *testing.T) {
	parsed := loadTestRobots(t)

	expected := &RobotsData{
		Groups: []RobotsGroup{
			RobotsGroup{
				UserAgents: []string{"a"},
				Disallow:   []string{"/a", "/b", "/c"},
			},
			RobotsGroup{
				UserAgents: []string{"b", "c"},
				Disallow:   []string{"/foo/bar"},
			},
			RobotsGroup{
				UserAgents: []string{"*"},
				Disallow:   []string{"/"},
			},
		},
	}

	compareParsedRobots(parsed, expected, t)
}

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func compareParsedRobots(a, b *RobotsData, t *testing.T) {
	if len(a.Groups) != len(b.Groups) {
		t.Errorf("a.Groups != b.Groups: %d != %d", len(a.Groups), len(b.Groups))
	}
	for i := 0; i < intMin(len(a.Groups), len(b.Groups)); i++ {
		compareGroup(a.Groups[i], b.Groups[i], t)
	}
}

func compareGroup(a, b RobotsGroup, t *testing.T) {
	if len(a.UserAgents) != len(b.UserAgents) {
		t.Errorf("UserAgents lists not equal: %s %s", a.UserAgents, b.UserAgents)
	} else {
		for i := range a.UserAgents {
			if a.UserAgents[i] != b.UserAgents[i] {
				t.Errorf("UserAgent %s != %s", a.UserAgents[i], b.UserAgents[i])
			}
		}
	}
	if len(a.Disallow) != len(b.Disallow) {
		t.Errorf("Disallow lists not equal: %s %s", a.Disallow, b.Disallow)
	} else {
		for i := range a.Disallow {
			if a.Disallow[i] != b.Disallow[i] {
				t.Errorf("Disallow %s != %s", a.Disallow[i], b.Disallow[i])
			}
		}
	}
}

func TestGetRobotsForURL(t *testing.T) {
	testFile, err := os.Open("testdata/test_robots.txt")
	if err != nil {
		t.Fatalf("Could not open testdata: %s", err)
	}
	defer testFile.Close()
	resp := &http.Response{
		StatusCode: 200,
		Body:       testFile,
	}
	factory := &mock.MockClientFactory{
		ForeverClient: &mock.MockClient{
			NextResponse: resp,
		},
	}
	data, err := GetRobotsForURL(&url.URL{Scheme: "http", Host: "localhost"}, factory)
	if err != nil {
		t.Errorf("Error when loading robots for URL: %v", err)
	}
	if data == nil {
		t.Error("Got nil data when loading robots for URL.")
	}
	exp := "http://localhost/robots.txt"
	if visited := factory.ForeverClient.Requests[0].String(); visited != exp {
		t.Errorf("Expected to request %s, visited %s", exp, visited)
	}
}

func TestGetForUserAgent_A(t *testing.T) {
	parsed := loadTestRobots(t)
	paths := parsed.GetForUserAgent("a")
	for i, p := range []string{"/a", "/b", "/c"} {
		if paths[i] != p {
			t.Errorf("Expected path %s, got %s.", p, paths[i])
		}
	}
}

func TestGetForUserAgent_Wildcard(t *testing.T) {
	parsed := loadTestRobots(t)
	paths := parsed.GetForUserAgent("wildcard")
	if len(paths) != 1 {
		t.Fatalf("Invalid response length: %d", len(paths))
	}
	if paths[0] != "/" {
		t.Errorf("Expected restricting /, got %s.", paths[0])
	}
}

func TestGetForUserAgent_Wildcard_Empty(t *testing.T) {
	rd := &RobotsData{Groups: []RobotsGroup{}}
	if rv := rd.GetForUserAgent("*"); rv != nil {
		t.Errorf("Expected nil response for wildcard user-agent, but got %v.", rv)
	}
}

func TestGetAllPaths(t *testing.T) {
	parsed := loadTestRobots(t)
	expected := []string{"/a", "/b", "/c", "/foo/bar", "/"}
	results := parsed.GetAllPaths()
	if le, lr := len(expected), len(results); le != lr {
		t.Fatalf("Expected and result lengths differ: %d/%d", le, lr)
	}
	for i, p := range expected {
		if p != results[i] {
			t.Errorf("Expected path %s, got %s", p, results[i])
		}
	}
}
