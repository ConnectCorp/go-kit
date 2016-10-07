package server

import (
	"crypto/tls"
	"github.com/PuerkitoBio/rehttp"
	"github.com/smartystreets/go-aws-auth"
	"net"
	"net/http"
	"net/url"
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

// AWSSigningHTTPTransport is an HTTP transport that signs AWS requests before sending them to the server.
type AWSSigningHTTPTransport struct {
	*http.Transport
}

// RoundTrip implements the http.RoundTripper interface.
func (a *AWSSigningHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	awsauth.Sign(req)
	return a.Transport.RoundTrip(req)
}

// MakeAWSSigningHTTPClient makes an http client that signs outgoing requests to AWS.
func MakeAWSSigningHTTPClient(retry rehttp.RetryFn) *http.Client {
	if retry == nil {
		retry = func(attempt rehttp.Attempt) bool {
			return false
		}
	}

	return &http.Client{
		Transport: rehttp.NewTransport(
			&AWSSigningHTTPTransport{
				Transport: &http.Transport{
					// Note that this ignores environment proxy settings for security reasons.
					Dial: (&net.Dialer{
						Timeout:   defaultDialerTimeout,
						KeepAlive: defaultDialerKeepAliveTimeout,
					}).Dial,
					TLSHandshakeTimeout:   defaultTLSHandshakeTimeout,
					ResponseHeaderTimeout: defaultResponseHeaderTimeout,
					ExpectContinueTimeout: defaultExpectContinueTimeout,
				},
			},
			retry,
			rehttp.ExpJitterDelay(defaultBaseExpJitterDelay, defaultMaxExpJitterDelay)),
	}
}

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

// mockTransport is an HTTP transport that redirects all requests to the given URL without using a proxy.
type mockTransport struct {
	mockServerURL *url.URL
	*http.Transport
}

// RountTrip implements the RoundTripper interface.
func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = m.mockServerURL.Scheme
	req.URL.Host = m.mockServerURL.Host
	return m.Transport.RoundTrip(req)
}

// MakeMockServerHTTPClient makes an HTTP client that redirects all requests to the given MockServer URL.
func MakeMockServerHTTPClient(mockServerURL *url.URL) *http.Client {
	return &http.Client{
		Transport: &mockTransport{
			mockServerURL: mockServerURL,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					// This is needed because the mock service is indeed a MITM attack.
					// Since this is enabled only for testing, we are cool with it.
					InsecureSkipVerify: true,
				},
			},
		},
	}
}

// MakeHTTPClientForConfig makes the HTTP client based on the TestProxy config value.
func MakeHTTPClientForConfig(config *CommonConfig, retry rehttp.RetryFn) *http.Client {
	if config.TestProxy.URL != nil {
		return MakeTestHTTPClient(config.TestProxy.URL)
	}
	return MakeProdHTTPClient(retry)
}

// MakeHTTPClientForConfigUsingMockServer makes the HTTP client based on the TestMock config value.
func MakeHTTPClientForConfigUsingMockServer(config *CommonConfig, retry rehttp.RetryFn) *http.Client {
	if config.TestMock.URL != nil {
		return MakeMockServerHTTPClient(config.TestMock.URL)
	}
	return MakeProdHTTPClient(retry)
}
