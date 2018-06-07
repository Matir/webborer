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
	"strings"
)

// StringSliceFlag is a flag.Value that takes a comma-separated string and turns
// it into a slice of strings.
type StringSliceFlag []string

// Satisfies flag.Value interface and splits value on commas
func (f *StringSliceFlag) String() string {
	if f == nil {
		return ""
	}
	return strings.Join(*f, ",")
}

func (f *StringSliceFlag) Set(value string) error {
	*f = append(*f, strings.Split(value, ",")...)
	return nil
}
