package main

import (
	"net/url"
	"strings"
)

// An Expander is responsible for taking input URLs and expanding them to
// include all of the words in the wordlist.
type Expander struct {
	// List of words to expand
	Wordlist *[]string
	// Function to count new instances
	Adder QueueAddCount
}

// Update the wordlist to contain directory & non-directory entries
func (e *Expander) ProcessWordlist() {
	newList := make([]string, 0)
	for _, w := range *e.Wordlist {
		newList = append(newList, w)
		if strings.Contains(w, ".") {
			continue
		}
		if w[len(w)-1] == byte('/') {
			continue
		}
		newList = append(newList, w+"/")
	}
	e.Wordlist = &newList
}

func (E *Expander) Expand(in <-chan *url.URL) <-chan *url.URL {
	out := make(chan *url.URL, cap(in))
	go func() {
		for e := range in {
			out <- e
			E.Adder(len(*E.Wordlist))
			for _, word := range *E.Wordlist {
				out <- ExtendURL(e, word)
			}
		}
		close(out)
	}()

	return out
}

func ExtendURL(u *url.URL, tail string) *url.URL {
	extended := *u
	if !URLIsDir(u) {
		extended.Path += "/" + tail
	} else {
		extended.Path += tail
	}
	return &extended
}
