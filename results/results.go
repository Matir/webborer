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

// Package results provides writers for several types of output.
package results

import (
	"encoding/csv"
	"fmt"
	ss "github.com/Matir/webborer/settings"
	"github.com/Matir/webborer/task"
	"io"
	"net/http"
	"net/url"
	"os"
)

// Types of links
type LinkType int

const (
	LinkHREF = LinkType(iota)
	LinkIMG
	LinkScript
	LinkStyle
	LinkUnknown
)

var LinkTypes = []string{
	"href",
	"img",
	"script",
	"style",
	"",
}

// This is the result emitted by the worker for each URL tested.
type Result struct {
	// URL of resource
	URL *url.URL
	// Host header (if different)
	Host string
	// HTTP Status Code
	Code int
	// Error if one occurred
	Error error
	// Redirect URL
	Redir *url.URL
	// Content length
	Length int64
	// Content-type header
	ContentType string
	// Known Headers
	RequestHeader http.Header
	// Response headers
	ResponseHeader http.Header
	// Group used for potentially bucketing results
	ResultGroup string
	// Links contained in result
	Links map[string]LinkType
}

// Create a new result.
func NewResult(URL *url.URL, host string) *Result {
	rv := &Result{
		URL:  URL,
		Host: host,
	}
	rv.ResultGroup = GetResultGroup(rv)
	return rv
}

func NewResultForTask(t *task.Task) *Result {
	rv := NewResult(t.URL, t.Host)
	rv.RequestHeader = t.Header
	return rv
}

type ResultGroupGenerator func(*Result) string

var GetResultGroup ResultGroupGenerator = func(*Result) string { return "" }

// Convert a result to a minimal string
func (r *Result) String() string {
	var host string
	if r.Host != "" {
		host = fmt.Sprintf(" (%s)", r.Host)
	}
	return fmt.Sprintf(
		"%s%s: %d",
		r.URL.String(),
		host,
		r.Code)
}

// Add a link to these results.
func (r *Result) AddLink(URL *url.URL, ltype LinkType) {
	if r.Links == nil {
		r.Links = make(map[string]LinkType)
	}
	r.Links[URL.String()] = ltype
}

// ResultsManager provides an interface for reading results from a channel and
// writing them to some form of output.
type ResultsManager interface {
	// Run reads all of the Results in the given channel and writes them to an
	// appropriate output sink.  Run should start its own goroutine for the bulk
	// of the work.
	Run(<-chan *Result)
	// Wait until the channel has been read and output done.
	Wait()
}

type baseResultsManager struct {
	finished chan bool
}

// Available output formats as strings.
var OutputFormats = []string{"text", "csv", "html", "diff"}

func init() {
	ss.SetOutputFormats(OutputFormats)
}

// Returns true if this is a "useful" result
func FoundSomething(code int) bool {
	return (code != 0 &&
		code != http.StatusNotFound &&
		code != http.StatusGone &&
		code != http.StatusBadGateway &&
		code != http.StatusServiceUnavailable &&
		code != http.StatusGatewayTimeout)
}

// Returns true if this result should be included in reports
func ReportResult(res *Result) bool {
	return res.Error == nil && FoundSomething(res.Code)
}

// Construct a ResultsManager for the given settings in the ss.ScanSettings.
// Returns an object satisfying the ResultsManager interface or an error.
func GetResultsManager(settings *ss.ScanSettings) (ResultsManager, error) {
	var writer io.WriteCloser
	var fp *os.File
	var err error

	format := settings.OutputFormat
	if settings.OutputPath == "" {
		writer = os.Stdout
	} else {
		if fp, err = os.Create(settings.OutputPath); err != nil {
			return nil, err
		} else {
			writer = fp
		}
	}

	if settings.RunMode == ss.RunModeLinkCheck {
		rm := &LinkCheckResultsManager{writer: writer, fp: fp, format: format, baseURL: settings.BaseURLs[0]}
		if err := rm.init(); err != nil {
			return nil, err
		}
		return rm, nil
	}

	switch {
	case format == "text":
		return &PlainResultsManager{writer: writer, fp: fp, redirs: settings.IncludeRedirects}, nil
	case format == "csv":
		return &CSVResultsManager{writer: csv.NewWriter(writer), fp: fp}, nil
	case format == "html":
		// TODO: do more than the first BaseURL
		return &HTMLResultsManager{writer: writer, fp: fp, BaseURL: settings.BaseURLs[0]}, nil
	case format == "diff":
		GetResultGroup = func(r *Result) string { return r.URL.Host }
		return NewDiffResultsManager(writer), nil
	}

	return nil, fmt.Errorf("Invalid output type: %s", format)
}

func (b *baseResultsManager) start() {
	b.finished = make(chan bool)
}

func (b *baseResultsManager) done() {
	b.finished <- true
}

func (b *baseResultsManager) Wait() {
	<-b.finished
}
