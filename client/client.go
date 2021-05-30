// Copyright 2017 Google Inc. All Rights Reserved.
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

// Package client provides a client for making HTTP requests.
package client

import (
	"encoding/base64"
	"fmt"
	"github.com/Matir/webborer/logging"
	"net/http"
	"net/url"
	"strings"
)

// Client is a thin wrapper around http.Client to make enhancements to
// support our use case.
type Client interface {
	RequestURL(*url.URL) (*http.Response, error)
	Request(*url.URL, string, string, http.Header) (*http.Response, error)
	SetCheckRedirect(func(*http.Request, []*http.Request) error)
}

// This interface just allows us to substitute a mock in tests
type httpClientInt interface {
	Do(req *http.Request) (*http.Response, error)
}

// httpClient is the concrete implementation of the Client interface.
//
// The Client member is almost always an http.Client outside of tests.
type httpClient struct {
	Client       httpClientInt
	UserAgent    string
	HTTPUsername string
	HTTPPassword string
	basicAuthStr string
}

// Request the URL given.
//
// Handles HTTP Authentication & Custom Headers
func (c *httpClient) RequestURL(u *url.URL) (*http.Response, error) {
	logging.Infof("Deprectated function RequestURL is called.")
	return c.Request(u, "", "GET", nil)
}

// Request the URL given with optional overrides.
//
// Handles HTTP Authentication & Custom Headers
func (c *httpClient) Request(u *url.URL, host, method string, header http.Header) (*http.Response, error) {
	req := c.makeRequest(u, method, host, header)
	resp, err := c.Client.Do(req)
	if err != nil {
		return resp, err
	}
	// Check if we can bypass auth checks via X-Forwarded-For & X-Real-IP
        if (resp.StatusCode>= 400) && (resp.StatusCode <600){
                switch resp.StatusCode{
                case 401,403,504,511:
                        req = c.makeRequest(u, method)
                        req.Header.Add("X-Forwarded-For","127.0.0.1")
                        req.Header.Add("X-Real-IP" ,"127.0.0.1")
                        resp_temp, err := c.Client.Do(req)
                        if (err == nil) &&  (resp_temp.StatusCode>=200) && (resp_temp.StatusCode<400){

                                logging.Logf(logging.LogWarning,"Pontential AUTH BYPASS in " + u.String() + " via X-Forwarded-For/X-Real-IP: 127.0.0.1")
                        }

                default:

                }
        }
	// Handle an authentication required response
	if resp.StatusCode == 401 {
		authHeader := resp.Header.Get("WWW-Authenticate")
		// No request for auth
		if authHeader == "" {
			return resp, nil
		}
		// No U/P available
		if c.HTTPUsername == "" && c.HTTPPassword == "" {
			return resp, nil
		}
		req = c.makeRequest(u, method, host, header)
		err = c.addAuthHeader(req, authHeader)
		if err != nil {
			logging.Logf(logging.LogInfo, err.Error())
			return resp, nil
		}
		resp, err = c.Client.Do(req)
		if err != nil {
			return resp, err
		}
	}
	return resp, nil
}

// Build a request with our preferred options
func (c *httpClient) makeRequest(u *url.URL, method, host string, header http.Header) *http.Request {
	req, _ := http.NewRequest(method, u.String(), nil)
	req.Host = host
	if header != nil {
		req.Header = header
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	return req
}

func (c *httpClient) SetCheckRedirect(checker func(*http.Request, []*http.Request) error) {
	cli, ok := c.Client.(*http.Client)
	if !ok {
		logging.Logf(logging.LogError, "Unable to set CheckRedirect, type assertion failed.")
		return
	}
	cli.CheckRedirect = checker
}

// Add an authentication header in response to authHeader
func (c *httpClient) addAuthHeader(req *http.Request, authHeader string) error {
	pieces := strings.SplitN(authHeader, " ", 2)
	if strings.ToLower(pieces[0]) == "basic" {
		req.Header.Add("Authorization", "Basic "+c.getBasicAuthStr())
		return nil
	}
	return fmt.Errorf("Unsupported WWW-Authenticate Method: %s", pieces[0])
}

// Build the base64-encoded username/password string
func (c *httpClient) getBasicAuthStr() string {
	if c.basicAuthStr != "" {
		return c.basicAuthStr
	}
	userpass := c.HTTPUsername + ":" + c.HTTPPassword
	c.basicAuthStr = base64.StdEncoding.EncodeToString([]byte(userpass))
	return c.basicAuthStr
}
