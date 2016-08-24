package server

import (
	"crypto/tls"
	"github.com/PuerkitoBio/rehttp"
	"net"
	"net/http"
	"net/url"
	"time"
	"github.com/prometheus/common/config"
)

const (
	defaultDialerTimeout          = 5 * time.Second
	defaultDialerKeepAliveTimeout = 30 * time.Second
	defaultTLSHandshakeTimeout    = 10 * time.Second
	defaultResponseHeaderTimeout  = 30 * time.Second
	defaultExpectContinueTimeout  = 1 * time.Second
	defaultMaxRetries             = 3
	defaultBaseExpJitterDelay     = 100 * time.Millisecond
	defaultMaxExpJitterDelay      = 5 * time.Second
)

// MakeProdHTTPClient makes an HTTP client suitable for use in production.
func MakeProdHTTPClient(retry rehttp.RetryFn) *http.Client {
	if retry == nil {
		retry = rehttp.RetryAll(
			rehttp.RetryMaxRetries(defaultMaxRetries),
			rehttp.RetryAny(
				rehttp.RetryTemporaryErr(),
				rehttp.RetryStatusInterval(500, 600),
			),
		)
	}

	return &http.Client{
		Transport: rehttp.NewTransport(
			&http.Transport{
				// Note that this ignores environment proxy settings for security reasons.
				Dial: (&net.Dialer{
					Timeout:   defaultDialerTimeout,
					KeepAlive: defaultDialerKeepAliveTimeout,
				}).Dial,
				TLSHandshakeTimeout:   defaultTLSHandshakeTimeout,
				ResponseHeaderTimeout: defaultResponseHeaderTimeout,
				ExpectContinueTimeout: defaultExpectContinueTimeout,
			},
			retry,
			rehttp.ExpJitterDelay(defaultBaseExpJitterDelay, defaultMaxExpJitterDelay)),
	}
}

// MakeTestHTTPClient makes an HTTP client suitable for use in test environments.
func MakeTestHTTPClient(testProxyURL *url.URL) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(testProxyURL),
			TLSClientConfig: &tls.Config{
				// This is needed because the test proxy is indeed a MITM attack.
				// Since this is enabled only for testing, we are cool with it.
				InsecureSkipVerify: true,
			},
		},
	}
}

// MakeHTTPClientForConfig makes the HTTP client based on the TestProxy config value.
func MakeHTTPClientForConfig(config *CommonConfig, retry rehttp.RetryFn) *http.Client {
	return MakeHTTPClientForConfigValue(config.TestProxy.URL, retry)
}

// MakeHTTPClientForConfig makes a test HTTP client if testURL is not nil, a prod HTTP client if it is nil.
func MakeHTTPClientForConfigValue(testURL *url.URL, retry rehttp.RetryFn) *http.Client {
	if testURL != nil {
		return MakeTestHTTPClient(testURL)
	}
	return MakeProdHTTPClient(retry)
}
