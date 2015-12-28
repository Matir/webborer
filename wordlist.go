package main

import (
	"bufio"
	"io"
	"os"
	"strings"
)

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
func LoadDefaultWordlist() ([]string, error) {
	return ReadWordlist(strings.NewReader(DefaultWordlist))
}
