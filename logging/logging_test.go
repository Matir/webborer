// Copyright 2017 Google Inc. All Rights Reserved.
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

package logging

import (
	"io/ioutil"
	"log"
	"testing"
)

func nullLog() {
	defaultLogger = log.New(ioutil.Discard, "", log.Ldate|log.Ltime|log.Lshortfile)
}

func TestLogLevelStrings(t *testing.T) {
	if len(LogLevelStrings) != logLevelMax {
		t.Errorf("Incorrect log level strings: %d != %d.", len(LogLevelStrings), logLevelMax)
	}
}

func TestResetLog(_ *testing.T) {
	// No-op
	ResetLog("", "")
	// Set both
	ResetLog("/dev/stderr", "WARNING")
}

func TestLogf(_ *testing.T) {
	nullLog()
	Logf(LogDebug, "Test Logf.")
	Logf(LogFatal, "Testing... %v", true)
}

func TestLogLevels(_ *testing.T) {
	nullLog()
	Debugf("Test %s", "Debugf")
	Infof("Test %s", "Infof")
	Warningf("Test %s", "Warningf")
	Errorf("Test %s", "Errorf")
	Fatalf("Test %s", "Fatalf")
}
