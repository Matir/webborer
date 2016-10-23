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
	"testing"
	"time"
)

func TestNewProxyClientFactory_Default(t *testing.T) {
	if fac, err := NewProxyClientFactory([]string{}, time.Nanosecond, ""); err != nil {
		t.Errorf("Unable to construct empty proxy client factory.")
	} else if fac == nil {
		t.Errorf("Returned nil factory on default.")
	}
}

func TestNewProxyClientFactory_SingleProxy(t *testing.T) {
	if fac, err := NewProxyClientFactory([]string{"socks5://localhost"}, time.Nanosecond, ""); err != nil {
		t.Errorf("Unable to construct empty proxy client factory.")
	} else if fac == nil {
		t.Errorf("Returned nil factory on single proxy.")
	}
}

func TestNewProxyClientFactory_UnsupportedMethod(t *testing.T) {
	proxies := []string{"socks5://localhost", "foo://localhost"}
	if fac, err := NewProxyClientFactory(proxies, time.Nanosecond, ""); err == nil {
		t.Errorf("Expected error parsing protocol.")
	} else if fac != nil {
		t.Errorf("Expected nil factory with invalid protocol.")
	}
}

func TestNewProxyClientFactory_InvalidURL(t *testing.T) {
	proxies := []string{"://"}
	if fac, err := NewProxyClientFactory(proxies, time.Nanosecond, ""); err == nil {
		t.Errorf("Expected error parsing URL")
	} else if fac != nil {
		t.Errorf("Expected nil factory with invalid URL.")
	}
}

func TestNewProxyClientFactory_NoHost(t *testing.T) {
	proxies := []string{"socks5://"}
	if fac, err := NewProxyClientFactory(proxies, time.Nanosecond, ""); err == nil {
		t.Errorf("Expected error parsing URL")
	} else if fac != nil {
		t.Errorf("Expected nil factory with missing host.")
	}
}

func TestPCFGet_NoProxies(t *testing.T) {
	fac, _ := NewProxyClientFactory([]string{}, time.Nanosecond, "")
	cli := fac.Get()
	if cli == nil {
		t.Errorf("Got nil client for no proxies.")
	}
}

func TestPCFGet_SingleProxy(t *testing.T) {
	fac, _ := NewProxyClientFactory([]string{"socks5://localhost"}, time.Nanosecond, "")
	cli := fac.Get()
	if cli == nil {
		t.Errorf("Got nil client for one proxy.")
	}
}

func TestPCFGet_TwoProxies(t *testing.T) {
	fac, _ := NewProxyClientFactory([]string{"socks5://localhost", "socks4://localhost:9000"}, time.Nanosecond, "")
	cli := fac.Get()
	if cli == nil {
		t.Errorf("Got nil client for two proxies.")
	}
}
