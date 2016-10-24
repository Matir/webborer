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

package mock

import (
	"bytes"
	"errors"
	"github.com/Matir/gobuster/client"
	"io/ioutil"
	"net/http"
	"net/url"
)

type MockClientFactory struct {
	NextClient *MockClient
}
type MockClient struct {
	NextResponse *http.Response
}

func (f *MockClientFactory) Get() client.Client {
	if f.NextClient != nil {
		c := f.NextClient
		f.NextClient = nil
		return c
	}
	return &MockClient{}
}

func (c *MockClient) RequestURL(u *url.URL) (*http.Response, error) {
	if c.NextResponse == nil {
		return nil, errors.New("No NextResponse for MockClient.")
	}
	r := c.NextResponse
	c.NextResponse = nil
	return r, nil
}

func (_ *MockClient) SetCheckRedirect(_ func(*http.Request, []*http.Request) error) {}

func ResponseFromString(s string) *http.Response {
	cb := ioutil.NopCloser(bytes.NewBufferString(s))
	return &http.Response{
		Body: cb,
	}
}

func MockRobotsResponse() *http.Response {
	s := `User-agent: *
Disallow: /a`
	return ResponseFromString(s)
}
