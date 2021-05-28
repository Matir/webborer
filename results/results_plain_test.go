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
	"strings"
	"testing"
)

// TODO: refactor this test to have a single test runner
func TestPlainResultsManager_Basic(t *testing.T) {
	buf := bytes.Buffer{}
	mgr := &PlainResultsManager{
		writer: &buf,
		redirs: true,
	}
	rchan := make(chan *Result)
	mgr.Run(rchan)
	for _, r := range makeTestResults() {
		rchan <- r
	}
	close(rchan)
	mgr.Wait()
	lines := strings.Split(buf.String(), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines of output, got %d", len(lines))
	}
}
