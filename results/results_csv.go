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
	"encoding/csv"
	"fmt"
	"net/url"
	"os"
)

// CSVResultsManager writes a CSV containing all of the results.
type CSVResultsManager struct {
	baseResultsManager
	writer *csv.Writer
	fp     *os.File
}

func (rm *CSVResultsManager) Run(res <-chan *Result) {
	go func() {
		rm.start()
		defer func() {
			rm.writer.Flush()
			if rm.fp != nil {
				rm.fp.Close()
			}
			rm.done()
		}()

		// Header line
		rm.writer.Write([]string{"code", "url", "content_length", "redirect_url"})

		for r := range res {
			rm.runOne(r)
		}
	}()
}

func (rm *CSVResultsManager) runOne(res *Result) {
	if !ReportResult(res) {
		return
	}
	var clen string
	if res.Length >= 0 {
		clen = fmt.Sprintf("%d", res.Length)
	}
	record := []string{
		fmt.Sprintf("%d", res.Code),
		res.URL.String(),
		clen,
		maybeStringURL(res.Redir),
	}
	rm.writer.Write(record)
}

func maybeStringURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	return u.String()
}
