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

package wordlist

import (
	"testing"
)

func TestLoadBuiltinWordlist(t *testing.T) {
	for _, wl := range []string{"default", "short"} {
		if list, err := LoadBuiltinWordlist(wl); err != nil {
			t.Errorf("Error when loading builtin wordlist %s: %v", wl, err)
		} else if list == nil {
			t.Errorf("Expected non-nil wordlist for %s.", wl)
		} else if len(list) == 0 {
			t.Errorf("No error, but builtin wordlist %s has len 0.", wl)
		}
	}

	if list, err := LoadBuiltinWordlist("yeah-not-real"); err == nil {
		t.Errorf("Expected error when loading non-existent wordlist.")
	} else if list != nil {
		t.Errorf("Expect nil wordlist for non-existent request.")
	}
}
