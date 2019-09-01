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

package client

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// Mock httpClient that returns arbitrary responses
type mockHttpClient struct {
	resps []*http.Response
	err   error
}

func makeMockHttpClient(resps ...*http.Response) *mockHttpClient {
	return &mockHttpClient{resps: resps}
}

func (c *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	if len(c.resps) == 0 {
		if c.err != nil {
			return nil, c.err
		}
		return nil, nil
	}
	resp, left := c.resps[0], c.resps[1:]
	c.resps = left
	if resp != nil {
		resp.Request = req
	}
	if c.err != nil {
		return resp, c.err
	}
	return resp, nil
}

// Mock httpClient that checks auth with "user" and "pass"
type mockAuthHttpClient struct {
	firstDone bool
}

func (c *mockAuthHttpClient) Do(req *http.Request) (*http.Response, error) {
	//TODO: make this method return different responses based on where it fails
	if !c.firstDone {
		resp := &http.Response{
			StatusCode: 401,
			Header:     make(http.Header, 0),
		}
		resp.Header.Set("WWW-Authenticate", "Basic realm=\"testing\"")
		c.firstDone = true
		return resp, nil
	}
	authFailed := &http.Response{StatusCode: 401}
	auth := req.Header.Get("Authorization")
	if auth == "" {
		return authFailed, nil
	}
	pieces := strings.Split(auth, " ")
	method, token := pieces[0], pieces[1]
	if method != "Basic" {
		return authFailed, nil
	}
	up, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return authFailed, nil
	}
	if string(up) != "user:pass" {
		return authFailed, nil
	}
	resp := &http.Response{
		StatusCode: 200,
	}
	return resp, nil
}

// Actual tests begin here
func TestMakeRequest_Basic(t *testing.T) {
	c := &httpClient{}
	u := &url.URL{Scheme: "http", Host: "localhost", Path: "/"}
	req := c.makeRequest(u, "GET", "", nil)
	if req.URL.String() != u.String() {
		t.Errorf("URL does not match requested: %s != %s", req.URL.String(), u.String())
	}
}

func TestSetCheckRedirect(_ *testing.T) {
	c := &httpClient{Client: &http.Client{}}
	c.SetCheckRedirect(func(_ *http.Request, _ []*http.Request) error { return nil })
	c = &httpClient{Client: &mockHttpClient{}}
	c.SetCheckRedirect(func(_ *http.Request, _ []*http.Request) error { return nil })
}

// Basic test of the full client stack
func TestRequestURL_Basic(t *testing.T) {
	mockResp := &http.Response{
		StatusCode: 200,
	}
	mockClient := makeMockHttpClient(mockResp)
	c := &httpClient{Client: mockClient, HTTPUsername: "user", HTTPPassword: "pass"}
	u := &url.URL{Scheme: "http", Host: "localhost", Path: "/"}
	resp, err := c.RequestURL(u)
	if err != nil {
		t.Errorf("Got error: %v", err)
	}
	if resp == nil {
		t.Errorf("Got nil response!")
	}
	if resp.StatusCode != 200 {
		t.Errorf("Got non-200 response code!")
	}
}

// Test with HTTP Basic Auth
func TestRequestURL_BasicAuth(t *testing.T) {
	mockClient := &mockAuthHttpClient{}
	c := &httpClient{Client: mockClient, HTTPUsername: "user", HTTPPassword: "pass"}
	u := &url.URL{Scheme: "http", Host: "localhost", Path: "/"}
	resp, err := c.RequestURL(u)
	if err != nil {
		t.Errorf("Got error: %v", err)
	}
	if resp == nil {
		t.Errorf("Got nil response!")
	}
	if resp.StatusCode != 200 {
		t.Errorf("Got non-200 response code: %d", resp.StatusCode)
	}
}

// Test with access denied
func TestRequestURL_BasicAuth_NoAuthHeader(t *testing.T) {
	mockResp := &http.Response{
		StatusCode: 401,
	}
	mockClient := makeMockHttpClient(mockResp)
	c := &httpClient{Client: mockClient, HTTPUsername: "user", HTTPPassword: "pass"}
	u := &url.URL{Scheme: "http", Host: "localhost", Path: "/"}
	resp, err := c.RequestURL(u)
	if err != nil {
		t.Errorf("Got error: %v", err)
	}
	if resp == nil {
		t.Errorf("Got nil response!")
	}
	if resp.StatusCode != 401 {
		t.Errorf("Got non-401 response code: %d", resp.StatusCode)
	}
}

// Test with HTTP Basic Auth, no password available
func TestRequestURL_BasicAuth_NoCreds(t *testing.T) {
	mockClient := &mockAuthHttpClient{}
	c := &httpClient{Client: mockClient}
	u := &url.URL{Scheme: "http", Host: "localhost", Path: "/"}
	resp, err := c.RequestURL(u)
	if err != nil {
		t.Errorf("Got error: %v", err)
	}
	if resp == nil {
		t.Errorf("Got nil response!")
	}
	if resp.StatusCode != 401 {
		t.Errorf("Got non-401 response code: %d", resp.StatusCode)
	}
}

// Test with digest
func TestRequestURL_DigestAuth(t *testing.T) {
	mockResp := &http.Response{
		StatusCode: 401,
		Header:     make(http.Header, 0),
	}
	// TODO: make this into a real digest header
	mockResp.Header.Set("WWW-Authenticate", "Digest realm=\"testing\"")
	mockClient := makeMockHttpClient(mockResp)
	c := &httpClient{Client: mockClient, HTTPUsername: "user", HTTPPassword: "pass"}
	u := &url.URL{Scheme: "http", Host: "localhost", Path: "/"}
	resp, err := c.RequestURL(u)
	if err != nil {
		t.Errorf("Got error: %v", err)
	}
	if resp == nil {
		t.Errorf("Got nil response!")
	}
	if resp.StatusCode != 401 {
		t.Errorf("Got non-401 response code: %d", resp.StatusCode)
	}
}
