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

package results

import (
	"github.com/Matir/webborer/settings"
	"net/url"
	"testing"
)

func makeTestResults() []Result {
	return []Result{
		Result{
			URL:         &url.URL{Scheme: "http", Host: "localhost", Path: "/"},
			Code:        200,
			ContentType: "text/html",
		},
		Result{
			URL:  &url.URL{Scheme: "http", Host: "localhost", Path: "/x"},
			Code: 404,
		},
		Result{
			URL:   &url.URL{Scheme: "http", Host: "localhost", Path: "/.git"},
			Code:  301,
			Redir: &url.URL{Scheme: "https", Host: "localhost", Path: "/.git"},
		},
	}

}

func TestGetResultsManager(t *testing.T) {
	for _, format := range OutputFormats {
		s := &settings.ScanSettings{OutputFormat: format, BaseURLs: []string{""}}
		if _, err := GetResultsManager(s); err != nil {
			t.Errorf("Unable to construct %s ResultsManager: %v", format, err)
		}
	}
}

func TestGetResultsManager_Invalid(t *testing.T) {
	s := &settings.ScanSettings{OutputFormat: "invalid"}
	if rm, err := GetResultsManager(s); err == nil {
		t.Error("Expecting error for invalid ResultsManager.")
	} else if rm != nil {
		t.Error("Expecting nil ResultsManager for invalid type.")
	}
}

func TestFoundSomething(t *testing.T) {
	if !FoundSomething(200) {
		t.Error("Expected 200 to be meaningful, but was not.")
	}
	if FoundSomething(404) {
		t.Error("Expected 404 not to be meaningful, but was.")
	}
}

func TestReportResult(t *testing.T) {
	r := Result{Code: 200}
	if !ReportResult(r) {
		t.Error("Expected to report a result of 200.")
	}
}

func TestBaseFunctions(_ *testing.T) {
	brm := &baseResultsManager{}
	brm.start()
	go brm.done()
	brm.Wait()
}
