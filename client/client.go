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

package client

import (
	"net/http"
	"net/url"
)

type Client struct {
	http.Client
	UserAgent string
}

func (c *Client) RequestURL(u *url.URL) (*http.Response, error) {
	req := c.MakeRequest(u)
	return c.Do(req)
}

func (c *Client) MakeRequest(u *url.URL) *http.Request {
	// TODO: support other methods
	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("User-Agent", c.UserAgent)
	return req
}
