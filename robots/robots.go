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

// Package robots provides support for parsing robots.txt files.
package robots

import (
	"bytes"
	"github.com/Matir/gobuster/client"
	"io/ioutil"
	"net/url"
)

type RobotsData struct {
	Groups []RobotsGroup
}

type RobotsGroup struct {
	UserAgents []string
	Disallow   []string
}

func ParseRobotsTxt(text []byte) (*RobotsData, error) {
	lines := bytes.Split(text, []byte{'\n'})
	robots := RobotsData{Groups: make([]RobotsGroup, 0)}
	curr_group := newRobotsGroup()
	agents_finished := false
	for _, line := range lines {
		line := trimSpaceAndComments(line)
		directive, value := splitLine(line)
		switch string(directive) {
		case "user-agent":
			if agents_finished {
				robots.Groups = append(robots.Groups, curr_group)
				curr_group = newRobotsGroup()
				agents_finished = false
			}
			curr_group.UserAgents = append(curr_group.UserAgents, string(value))
		case "disallow":
			agents_finished = true
			curr_group.Disallow = append(curr_group.Disallow, string(value))
		}
	}
	if len(curr_group.UserAgents) > 0 {
		robots.Groups = append(robots.Groups, curr_group)
	}
	// TODO: error checking
	return &robots, nil
}

func trimSpaceAndComments(line []byte) []byte {
	sections := bytes.Split(line, []byte{'#'})
	return bytes.TrimSpace(sections[0])
}

func splitLine(line []byte) ([]byte, []byte) {
	if !bytes.Contains(line, []byte{':'}) {
		return []byte{}, []byte{}
	}
	sections := bytes.SplitN(line, []byte{':'}, 2)
	return bytes.ToLower(bytes.TrimSpace(sections[0])), bytes.TrimSpace(sections[1])
}

func newRobotsGroup() RobotsGroup {
	return RobotsGroup{
		UserAgents: make([]string, 0),
		Disallow:   make([]string, 0),
	}
}

func GetRobotsForURL(target *url.URL, factory client.ClientFactory) (*RobotsData, error) {
	client := factory.Get()
	ref, _ := url.Parse("/robots.txt")
	robotsURL := target.ResolveReference(ref)
	resp, err := client.RequestURL(robotsURL)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ParseRobotsTxt(body)
}

func (data *RobotsData) GetForUserAgent(targetAgent string) []string {
	for _, group := range data.Groups {
		for _, agent := range group.UserAgents {
			if agent == targetAgent {
				return group.Disallow
			}
		}
	}

	if targetAgent == "*" {
		return nil
	}

	// Fallback to '*'
	return data.GetForUserAgent("*")
}

func (data *RobotsData) GetAllPaths() []string {
	results := make([]string, 0)
	for _, group := range data.Groups {
		results = append(results, group.Disallow...)
	}
	return results
}
