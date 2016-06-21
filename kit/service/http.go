package service

import (
	"github.com/PuerkitoBio/rehttp"
	"net"
	"net/url"
	"crypto/tls"
	"net/http"
	"time"
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
func MakeProdHTTPClient() *http.Client {
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
			rehttp.RetryAll(
				rehttp.RetryMaxRetries(defaultMaxRetries),
				rehttp.RetryAny(
					rehttp.RetryTemporaryErr(),
					rehttp.RetryStatusInterval(500, 600),
				),
			),
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

