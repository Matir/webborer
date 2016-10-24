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
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
)

// Long test to thoroughly test CSV writing.
func TestWriteCSV(t *testing.T) {
	rchan := make(chan Result)
	buf := bytes.Buffer{}
	mgr := CSVResultsManager{
		writer: csv.NewWriter(&buf),
	}
	res := makeTestResults()
	mgr.Run(rchan)
	for _, r := range res {
		rchan <- r
	}
	close(rchan)
	mgr.Wait()
	lines := strings.Split(buf.String(), "\n")
	if len(lines) != 4 {
		t.Fatalf("Expected 2 lines of output, got %d.", len(lines))
	}
	hdr := "code,url,content_length,redirect_url"
	if lines[0] != hdr {
		t.Errorf("Expected header \"%s\", got header \"%s\".", hdr, lines[0])
	}
	resStr := "200,http://localhost/,0,"
	if lines[1] != resStr {
		t.Errorf("Expected result string \"%s\", got result string \"%s\".", resStr, lines[1])
	}
	resStr = "301,http://localhost/.git,0,https://localhost/.git"
	if lines[2] != resStr {
		t.Errorf("Expected result string \"%s\", got result string \"%s\".", resStr, lines[1])
	}
}
