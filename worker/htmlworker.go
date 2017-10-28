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

package worker

import (
	"github.com/Matir/webborer/logging"
	"github.com/Matir/webborer/util"
	"github.com/Matir/webborer/workqueue"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	maxHTMLWorkerSize = 10 * 1024 * 1024
)

type HTMLWorker struct {
	// Function to add future work
	adder workqueue.QueueAddFunc
}

func NewHTMLWorker(adder workqueue.QueueAddFunc) *HTMLWorker {
	return &HTMLWorker{adder: adder}
}

// Work on this response
func (w *HTMLWorker) Handle(URL *url.URL, body io.Reader) {
	limitedBody := io.LimitReader(body, maxHTMLWorkerSize)
	links := w.GetLinks(limitedBody)
	foundURLs := make([]*url.URL, 0, len(links))
	for _, l := range links {
		u, err := url.Parse(l)
		if err != nil {
			logging.Logf(logging.LogInfo, "Error parsing URL (%s): %s", l, err.Error())
			continue
		}
		// TODO: use <base> tag
		resolved := URL.ResolveReference(u)
		foundURLs = append(foundURLs, resolved)
		// Include parents of the found URL.
		// Worker will remove duplicates
		foundURLs = append(foundURLs, util.GetParentPaths(resolved)...)
	}
	w.adder(foundURLs...)
}

// Check if this response can be handled by this worker
func (*HTMLWorker) Eligible(resp *http.Response) bool {
	ct := resp.Header.Get("Content-type")
	if strings.ToLower(ct) != "text/html" {
		return false
	}
	// ContentLength is often -1, indicating unknown, so we'll try to parse those
	return resp.ContentLength == -1 || (resp.ContentLength > 0 && resp.ContentLength < maxHTMLWorkerSize)
}

// Get the links for the body.
func (*HTMLWorker) GetLinks(body io.Reader) []string {
	tree, err := html.Parse(body)
	if err != nil {
		logging.Logf(logging.LogInfo, "Unable to parse HTML document: %s", err.Error())
		return nil
	}
	links := collectElementAttributes(tree, "a", "href")
	links = append(links, collectElementAttributes(tree, "img", "src")...)
	links = append(links, collectElementAttributes(tree, "script", "src")...)
	links = append(links, collectElementAttributes(tree, "style", "src")...)
	return util.DedupeStrings(links)
}

func getElementsByTagName(root *html.Node, name string) []*html.Node {
	results := make([]*html.Node, 0)
	var handleNode func(*html.Node)
	handleNode = func(node *html.Node) {
		if node.Type == html.ElementNode && strings.ToLower(node.Data) == name {
			results = append(results, node)
		}
		for n := node.FirstChild; n != nil; n = n.NextSibling {
			handleNode(n)
		}
	}
	handleNode(root)
	return results
}

func getElementAttribute(node *html.Node, attrName string) *string {
	for _, a := range node.Attr {
		if strings.ToLower(a.Key) == attrName {
			return &a.Val
		}
	}
	return nil
}

func collectElementAttributes(root *html.Node, tagName, attrName string) []string {
	results := make([]string, 0)
	for _, el := range getElementsByTagName(root, tagName) {
		if val := getElementAttribute(el, attrName); val != nil {
			results = append(results, *val)
		}
	}
	return results
}
