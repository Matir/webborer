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

package settings

import (
	"github.com/Matir/gobuster/logging"
	"testing"
)

func TestRobotsModeStrings(t *testing.T) {
	if len(robotsModeStrings) != robotsModeMax {
		t.Errorf("RobotsModeStrings != enum: %d vs %d", len(robotsModeStrings), robotsModeMax)
	}
}

// Test some defaults
func TestNewScanSettings(t *testing.T) {
	ss := NewScanSettings()
	if ss == nil {
		t.Fatalf("NewScanSettings returned nil!")
	}
	var foundLogLevel bool
	for _, l := range logging.LogLevelStrings {
		if l == ss.LogLevel {
			foundLogLevel = true
		}
	}
	if !foundLogLevel {
		t.Errorf("Invalid default loglevel: %s", ss.LogLevel)
	}
	if len(ss.Extensions) < 1 {
		t.Errorf("No default extensions!")
	}
	if len(ss.SpiderCodes) < 1 {
		t.Errorf("No HTTP codes to spider!")
	}
	if !ss.flagsSet {
		t.Errorf("Flags not initialized!")
	}
}
