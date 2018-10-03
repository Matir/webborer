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
	"github.com/matir/webborer/logging"
	"html/template"
	"io"
	"os"
)

// HTMLResultsManager writes an HTML file containing the results.
type HTMLResultsManager struct {
	baseResultsManager
	writer  io.Writer
	fp      *os.File
	BaseURL string
}

func (rm *HTMLResultsManager) Run(res <-chan *Result) {
	go func() {
		rm.start()
		rm.writeHeader()

		defer func() {
			rm.writeFooter()
			if rm.fp != nil {
				rm.fp.Close()
			}
			rm.done()
		}()

		for r := range res {
			if !ReportResult(r) {
				continue
			}
			if r.Redir != nil {
				continue
			}
			rm.writeResult(r)
		}
	}()
}

func (rm *HTMLResultsManager) writeHeader() {
	header := `{{define "HEAD"}}<html><head><title>webborer: {{.BaseURL}}</title></head><h2>Results for <a href="{{.BaseURL}}">{{.BaseURL}}</a></h2><table><tr><th>Code</th><th>URL</th><th>Size</th><th>Content-Type</th></tr>{{end}}`
	t, err := template.New("htmlResultsManager").Parse(header)
	if err != nil {
		logging.Logf(logging.LogWarning, "Error parsing a template: %s", err.Error())
	}
	data := struct {
		BaseURL string
	}{
		BaseURL: rm.BaseURL,
	}
	err = t.ExecuteTemplate(rm.writer, "HEAD", data)
	if err != nil {
		logging.Logf(logging.LogWarning, "Error writing template output: %s", err.Error())
	}
}

func (rm *HTMLResultsManager) writeFooter() {
	footer := `{{define "FOOTER"}}</table></html>{{end}}`
	t, err := template.New("htmlResultsManager").Parse(footer)
	if err != nil {
		logging.Logf(logging.LogWarning, "Error parsing a template: %s", err.Error())
	}
	err = t.ExecuteTemplate(rm.writer, "FOOTER", nil)
	if err != nil {
		logging.Logf(logging.LogWarning, "Error writing template output: %s", err.Error())
	}
}

func (rm *HTMLResultsManager) writeResult(res *Result) {
	// TODO: don't rebuild the template with each row
	tmpl := `{{define "ROW"}}<tr><td>{{.Code}}</td><td><a href="{{.URL.String}}">{{.URL.String}}</a></td><td>{{if ge .Length 0}}{{.Length}}{{end}}</td><td>{{.ContentType}}</td></tr>{{end}}`
	t, err := template.New("htmlResultsManager").Parse(tmpl)
	if err != nil {
		logging.Logf(logging.LogWarning, "Error parsing a template: %s", err.Error())
	}
	err = t.ExecuteTemplate(rm.writer, "ROW", res)
	if err != nil {
		logging.Logf(logging.LogWarning, "Error writing template output: %s", err.Error())
	}
}
