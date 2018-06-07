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
	"time"
)

// DurationFlag is a flag.Value that takes a Duration spec (see time.Duration)
// and parses it and stores the Duration.
type DurationFlag struct {
	d *time.Duration
}

// Satisfies flag.Value interface and converts to a duration based on seconds
func (f DurationFlag) String() string {
	if f.d == nil {
		return ""
	}
	return f.d.String()
}

func (f DurationFlag) Set(value string) error {
	if d, err := time.ParseDuration(value); err != nil {
		return err
	} else {
		*f.d = d
	}
	return nil
}
