// Copyright 2015 Google Inc. All Rights Reserved.
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
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
)

// First try loading from a file, then try loading from built-ins
func LoadWordlist(path string) ([]string, error) {
	if path == "" {
		return LoadBuiltinWordlist("default")
	}
	wl, wl_err := ReadWordlistFile(path)
	if wl_err == nil {
		return wl, nil
	}
	if wl, err := LoadBuiltinWordlist(path); err == nil {
		return wl, nil
	}
	return nil, wl_err
}

// Load a Wordlist from a file.
func ReadWordlistFile(path string) ([]string, error) {
	if fp, err := os.Open(path); err != nil {
		return nil, err
	} else {
		defer fp.Close()
		return ReadWordlist(fp)
	}
}

// Load a wordlist from a reader.
// This basically just splits the contents of a reader on newlines.
func ReadWordlist(rdr io.Reader) ([]string, error) {
	wordlist := make([]string, 0)
	scanner := bufio.NewScanner(rdr)
	for scanner.Scan() {
		w := string(scanner.Bytes())
		if w != "" {
			wordlist = append(wordlist, w)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return wordlist, nil
}

// Loads a built-in wordlist for basic scans.
func LoadBuiltinWordlist(which string) ([]string, error) {
	switch which {
	case "default":
		return ReadWordlist(strings.NewReader(DefaultWordlist))
	case "short":
		return ReadWordlist(strings.NewReader(ShortWordlist))
	}
	return nil, errors.New("No such built-in wordlist.")
}
