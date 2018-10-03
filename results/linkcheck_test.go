// Copyright 2018 Google Inc. All Rights Reserved.
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
	"net/http"
	"testing"
)

func TestInitLinkCheck(t *testing.T) {
	lcrm := &LinkCheckResultsManager{
		writer: &bytes.Buffer{},
	}
	if err := lcrm.init(); err == nil {
		t.Error("Expected error for missing format.")
	}
	lcrm.format = "html"
	if err := lcrm.init(); err != nil {
		t.Error("Did not expect an error.")
	}
	if _, ok := lcrm.writerImpl.(*linkCheckHTMLWriter); !ok {
		t.Error("Expected an HTML writer.")
	}
	if lcrm.resMap == nil {
		t.Error("Expected resMap to be initialized.")
	}
	lcrm.format = "csv"
	if err := lcrm.init(); err != nil {
		t.Error("Did not expect an error.")
	}
	if _, ok := lcrm.writerImpl.(*linkCheckCSVWriter); !ok {
		t.Error("Expected a CSV writer.")
	}
}

func TestCodeIsBroken(t *testing.T) {
	if codeIsBroken(http.StatusOK) {
		t.Error("StatusOK is not broken.")
	}
	if !codeIsBroken(http.StatusNotFound) {
		t.Error("StatusNotFound should be broken.")
	}
}
