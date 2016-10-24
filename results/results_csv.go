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

func (rm *CSVResultsManager) Run(res <-chan Result) {
	go func() {
		rm.start()
		defer func() {
			rm.writer.Flush()
			if rm.fp != nil {
				rm.fp.Close()
			}
			rm.done()
		}()

		maybeString := func(u *url.URL) string {
			if u == nil {
				return ""
			}
			return u.String()
		}

		rm.writer.Write([]string{"code", "url", "content_length", "redirect_url"})

		for r := range res {
			if !ReportResult(r) {
				continue
			}
			var clen string
			if r.Length >= 0 {
				clen = fmt.Sprintf("%d", r.Length)
			}
			record := []string{
				fmt.Sprintf("%d", r.Code),
				r.URL.String(),
				clen,
				maybeString(r.Redir),
			}
			rm.writer.Write(record)
		}
	}()
}
