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

// Package settings provides a central interface to webborer settings.
package settings

import (
	"fmt"
)

// Run mode specifies a couple of different ways for handling general
// operations.
type RunModeOption int

// We have a couple of different runmodes
const (
	RunModeEnumeration = iota
	RunModeDotProduct
)

var runModeStrings = [...]string{
	"enumeration",
	"dotproduct",
}

func (f *RunModeOption) String() string {
	if f == nil {
		return runModeStrings[RunModeEnumeration]
	}
	return runModeStrings[*f]
}

func (f *RunModeOption) Set(value string) error {
	for i, val := range runModeStrings {
		if val == value {
			*f = RunModeOption(i)
			return nil
		}
	}
	return fmt.Errorf("Unknown Run Mode: %s", value)
}
