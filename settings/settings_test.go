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
	"github.com/matir/webborer/logging"
	"testing"
	"time"
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

func TestStringSliceFlag(t *testing.T) {
	f := StringSliceFlag{}
	if f.String() != "" {
		t.Error("Expected empty string for empty StringSliceFlag.")
	}
	s := "a,b,c"
	if err := f.Set(s); err != nil {
		t.Errorf("Error when setting StringSliceFlag: %v", err)
	}
	if len(f) != 3 {
		t.Errorf("len(f) != 3, = %d", len(f))
	}
	if f.String() != s {
		t.Errorf("Differing strings: \"%s\" vs \"%s\".", f.String(), s)
	}
}

func TestIntSliceFlag(t *testing.T) {
	f := IntSliceFlag{}
	if f.String() != "" {
		t.Error("Expected empty string for empty IntSliceFlag.")
	}
	s := "1,2,3"
	if err := f.Set(s); err != nil {
		t.Errorf("Error when setting IntSliceFlag: %v", err)
	}
	if len(f) != 3 {
		t.Errorf("len(f) != 3, = %d", len(f))
	}
	if f.String() != s {
		t.Errorf("Differing strings: \"%s\" vs \"%s\".", f.String(), s)
	}
	if err := f.Set("xyz"); err == nil {
		t.Error("Expected error when setting invalid IntSliceFlag.")
	}
}

func TestDurationFlag_Empty(t *testing.T) {
	f := DurationFlag{}
	if f.String() != "" {
		t.Error("Expected empty string for empty DurationFlag.")
	}
}

func TestDurationFlag_String(t *testing.T) {
	d := time.Second
	f := DurationFlag{&d}
	if f.String() != "1s" {
		t.Errorf("Expected \"1s\" for duration: \"%s\"", f.String())
	}
}

func TestDurationFlag_Set_Valid(t *testing.T) {
	d := time.Duration(0)
	f := DurationFlag{&d}
	if err := f.Set("1s"); err != nil {
		t.Errorf("Error setting DurationFlag: %v", err)
	}
}

func TestDurationFlag_Set_Invalid(t *testing.T) {
	d := time.Duration(0)
	f := DurationFlag{&d}
	if err := f.Set("blah"); err == nil {
		t.Error("Expected error setting DurationFlag.")
	}
}

func TestRobotsFlag_Empty(t *testing.T) {
	f := RobotsModeOption(0)
	if f.String() != "ignore" {
		t.Errorf("Expected robots flag ignore, got %s.", f.String())
	}
}

func TestRobotsFlag_String(t *testing.T) {
	f := RobotsModeOption(0)
	if f.String() != "ignore" {
		t.Errorf("Expected robots flag ignore, got %s.", f.String())
	}
}

func TestRobotsFlag_Set_Valid(t *testing.T) {
	f := RobotsModeOption(0)
	if err := f.Set("obey"); err != nil {
		t.Errorf("Expected no error setting robots flag, got %v", err)
	}
	if f != ObeyRobots {
		t.Errorf("Expected flag to be %d, got %d.", ObeyRobots, f)
	}
}

func TestRobotsFlag_Set_Invalid(t *testing.T) {
	f := RobotsModeOption(0)
	if err := f.Set("wtfmate"); err == nil {
		t.Error("Expected error setting flag, got nil.")
	}
	if f != 0 {
		t.Errorf("Expected flag unchanged during error, got %d.", f)
	}
}

func TestScanSettings_String(t *testing.T) {
	ss := &ScanSettings{}
	if len(ss.String()) == 0 {
		t.Error("Expected string response, got nothing.")
	}
}

func TestScanSettings_GetScopes_Success(t *testing.T) {
	ss := &ScanSettings{
		BaseURLs: []string{"http://localhost/", "https://example.org"},
	}
	if scopes, err := ss.GetScopes(); err != nil {
		t.Errorf("Expected no error getting scope, got %v.", err)
	} else {
		if len(scopes) != len(ss.BaseURLs) {
			t.Errorf("Length mismatch: %d vs %d", len(scopes), len(ss.BaseURLs))
		}
		if scopes[0].String() != ss.BaseURLs[0] {
			t.Errorf("URL mismatch: %v vs %v", scopes[0], ss.BaseURLs[0])
		}
		if scopes[1].String() != ss.BaseURLs[1]+"/" {
			t.Errorf("URL mismatch: %v vs %v", scopes[1], ss.BaseURLs[1])
		}
	}
}

func TestScanSettings_GetScopes_Error(t *testing.T) {
	ss := &ScanSettings{
		BaseURLs: []string{"://localhost/"},
	}
	if scopes, err := ss.GetScopes(); err == nil {
		t.Error("Expected error, got nil.")
	} else if scopes != nil {
		t.Errorf("Expected nil scopes, got %v.", scopes)
	}
}

func TestScanSettings_Validate(t *testing.T) {
	ss := &ScanSettings{
		BaseURLs: []string{},
	}
	if err := ss.Validate(); err == nil {
		t.Errorf("Expected error with no BaseURLs.")
	}
	ss = &ScanSettings{
		BaseURLs: []string{"http://www.example.com"},
	}
	if err := ss.Validate(); err != nil {
		t.Errorf("Expected no errors with BaseURLs.")
	}
}
