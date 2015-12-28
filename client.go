package main

import (
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

var DefaultUserAgent = "GoBuster 0.01"

// A ClientFactory allows constructing HTTP Clients based on various Dialers or
// Transports.
type ClientFactory interface {
	Get() *http.Client
}

// ProxyClientFactory uses the h12.me/socks package to support SOCKS proxies
// when transporting requests to the webserver.
type ProxyClientFactory struct {
	proxyURLs []*url.URL
	timeout   time.Duration
}

// Create a ProxyClientFactory for the provided list of proxies.
func NewProxyClientFactory(proxies []string, timeout time.Duration) *ProxyClientFactory {
	factory := &ProxyClientFactory{timeout: timeout}
	for _, proxy := range proxies {
		u, err := url.Parse(proxy)
		if err != nil {
			Logf(LogWarning, "Unable to parse proxy: %s", proxy)
			continue
		}
		if _, ok := proxyTypeMap[u.Scheme]; !ok {
			Logf(LogWarning, "Invalid proxy protocol: %s", u.Scheme)
			continue
		}
		if u.Host == "" {
			Logf(LogWarning, "Missing host for proxy: %s", proxy)
			continue
		}
		factory.proxyURLs = append(factory.proxyURLs, u)
	}
	return factory
}

func (factory *ProxyClientFactory) Get() *http.Client {
	if len(factory.proxyURLs) == 0 {
		return &http.Client{Timeout: factory.timeout}
	}
	if len(factory.proxyURLs) == 1 {
		return clientForProxy(factory.proxyURLs[0], factory.timeout)
	}
	proxy := factory.proxyURLs[rand.Intn(len(factory.proxyURLs))]
	return clientForProxy(proxy, factory.timeout)
}

func clientForProxy(proxy *url.URL, timeout time.Duration) *http.Client {
	proto := proxyTypeMap[proxy.Scheme]
	dialer := socks.DialSocksProxy(proto, proxy.Host)
	cl := &http.Client{Transport: &http.Transport{Dial: dialer}, Timeout: timeout}
	return cl
}
