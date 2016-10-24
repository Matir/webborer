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
	"fmt"
	"io"
	"os"
)

// PlainResultsManager is designed to output a very basic output that is good
// for human reading, but not so good for machine parsing.  This is the default
// output and provides a decent way to review results on-screen.
type PlainResultsManager struct {
	baseResultsManager
	writer io.Writer
	fp     *os.File
	redirs bool
}

func (rm *PlainResultsManager) Run(res <-chan Result) {
	go func() {
		rm.start()
		defer func() {
			if rm.fp != nil {
				rm.fp.Close()
			}
			rm.done()
		}()

		for r := range res {
			if !ReportResult(r) {
				continue
			}
			if r.Redir == nil {
				if r.Length >= 0 {
					fmt.Fprintf(rm.writer, "%d %s (%d bytes)\n", r.Code, r.URL.String(), r.Length)
				} else {
					fmt.Fprintf(rm.writer, "%d %s\n", r.Code, r.URL.String())
				}
			} else if rm.redirs {
				fmt.Fprintf(rm.writer, "%d %s -> %s\n", r.Code, r.URL.String(), r.Redir.String())
			}
		}
	}()
}
