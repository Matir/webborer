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
	"github.com/matir/webborer/util"
	"net/http"
	"strings"
)

// HeaderFlag is an http.Header wrapped in the flag.Value interface.
type HeaderFlag http.Header

func (f *HeaderFlag) String() string {
	return util.StringHeader(http.Header(*f), " ")
}

func (f *HeaderFlag) Set(value string) error {
	if f == nil {
		panic("Nil HeaderFlag object in set!")
	}
	pieces := strings.SplitN(value, ":", 2)
	if len(pieces) != 2 {
		return fmt.Errorf("Header format is key: value")
	}
	key := strings.TrimSpace(pieces[0])
	val := strings.TrimSpace(pieces[1])
	http.Header(*f).Add(key, val)
	return nil
}

func (f *HeaderFlag) Header() http.Header {
	return http.Header(*f)
}
