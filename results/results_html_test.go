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
	"net/url"
	"testing"
)

// Very basic test, does not check output formatting.
func TestHTMLResultsManager_Basic(t *testing.T) {
	buf := bytes.Buffer{}
	mgr := &HTMLResultsManager{
		writer: &buf,
	}
	rchan := make(chan Result)
	res := Result{
		URL:  &url.URL{Scheme: "http", Host: "localhost", Path: "/"},
		Code: 200,
	}
	mgr.Run(rchan)
	rchan <- res
	close(rchan)
	mgr.Wait()
	if len(buf.String()) == 0 {
		t.Fatal("Expected some output, got nothing!")
	}
}
