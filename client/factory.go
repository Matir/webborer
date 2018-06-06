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

package client

import (
	"fmt"
	"github.com/matir/webborer/logging"
	"h12.me/socks"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

var proxyTypeMap = map[string]int{
	"socks":   socks.SOCKS4,
	"socks4":  socks.SOCKS4,
	"socks4a": socks.SOCKS4A,
	"socks5":  socks.SOCKS5,
}

// A ClientFactory allows constructing HTTP Clients based on various Dialers or
// Transports.
type ClientFactory interface {
	Get() Client
}

// ProxyClientFactory uses the h12.me/socks package to support SOCKS proxies
// when transporting requests to the webserver.
type ProxyClientFactory struct {
	proxyURLs    []*url.URL
	timeout      time.Duration
	userAgent    string
	httpUsername string
	httpPassword string
}

// Create a ProxyClientFactory for the provided list of proxies.
func NewProxyClientFactory(proxies []string, timeout time.Duration, agent string) (*ProxyClientFactory, error) {
	factory := &ProxyClientFactory{timeout: timeout, userAgent: agent}
	for _, proxy := range proxies {
		u, err := url.Parse(proxy)
		if err != nil {
			logging.Logf(logging.LogWarning, "Unable to parse proxy: %s", proxy)
			return nil, err
		}
		if _, ok := proxyTypeMap[u.Scheme]; !ok {
			logging.Logf(logging.LogWarning, "Invalid proxy protocol: %s", u.Scheme)
			return nil, fmt.Errorf("Invalid proxy protocol: %s", u.Scheme)
		}
		if u.Host == "" {
			logging.Logf(logging.LogWarning, "Missing host for proxy: %s", proxy)
			return nil, fmt.Errorf("Missing host for proxy: %s", proxy)
		}
		factory.proxyURLs = append(factory.proxyURLs, u)
	}
	return factory, nil
}

func (factory *ProxyClientFactory) SetUsernamePassword(username, password string) {
	factory.httpUsername = username
	factory.httpPassword = password
}

// Get a single client instance from the factory
func (factory *ProxyClientFactory) Get() Client {
	if len(factory.proxyURLs) == 0 {
		return &httpClient{
			Client:       &http.Client{Timeout: factory.timeout},
			UserAgent:    factory.userAgent,
			HTTPUsername: factory.httpUsername,
			HTTPPassword: factory.httpPassword,
		}
	}
	var cli *httpClient
	if len(factory.proxyURLs) == 1 {
		cli = clientForProxy(factory.proxyURLs[0], factory.timeout, factory.userAgent)
	} else {
		proxy := factory.proxyURLs[rand.Intn(len(factory.proxyURLs))]
		cli = clientForProxy(proxy, factory.timeout, factory.userAgent)
	}
	cli.HTTPUsername = factory.httpUsername
	cli.HTTPPassword = factory.httpPassword
	return cli
}

// Build a client for a particular proxy instance
func clientForProxy(proxy *url.URL, timeout time.Duration, agent string) *httpClient {
	proto := proxyTypeMap[proxy.Scheme]
	dialer := socks.DialSocksProxy(proto, proxy.Host)
	cl := &httpClient{
		Client: &http.Client{
			Transport: &http.Transport{
				Dial: dialer,
			},
			Timeout: timeout,
		},
		UserAgent: agent}
	return cl
}
