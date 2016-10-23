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
	"net/url"
	"testing"
)

func TestMakeRequest_Basic(t *testing.T) {
	c := &Client{}
	u := &url.URL{Scheme: "http", Host: "localhost", Path: "/"}
	req := c.MakeRequest(u)
	if req.URL.String() != u.String() {
		t.Errorf("URL does not match requested: %s != %s", req.URL.String(), u.String())
	}
}
