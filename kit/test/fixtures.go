package test

import (
	"golang.org/x/net/context"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/stretchr/testify/assert"
)

// MustNewRequest creates a new generic HTTP request for testing.
func MustNewRequest() *http.Request {
	req, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		panic(err)
	}
	return req
}

// TerminationMiddleware is a middleware that terminates processing with an error.
func TerminationMiddleware(_ context.Context, _ interface{}) (interface{}, error) {
	return nil, xerror.New("terminated")
}

// TestServer is a test HTTP server that validates incoming requests.
type TestServer struct {
	server *httptest.Server
	validator func(r *http.Request)
	responder func(w http.ResponseWriter, r *http.Request)
	received bool
}

// NewTestServer initialized a new TestServer.
func NewTestServer() *TestServer {
	cts := &TestServer{}
	cts.server = httptest.NewServer(cts)
	return cts
}

// URL return a URL to the test server with the given path.
func (cts *TestServer) URL(path string) string {
	return cts.server.URL+path
}

// SetValidator sets the validator function used for validating a request.
func (cts *TestServer) SetValidator(validator func(r *http.Request)) {
	cts.validator = validator
}

// SetResponder sets the responder function used for returning a response.
func (cts *TestServer) SetResponder(responder func(w http.ResponseWriter, r *http.Request)) {
	cts.responder = responder
}

// ServeHTTP implements the http.Handler interface.
func (cts *TestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cts.received = true
	if cts.validator != nil {
		cts.validator(r)
	}
	if cts.responder != nil {
		cts.responder(w, r)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

// AssertReceived asserts that a request has been received and resets the server.
func (cts *TestServer) AssertReceived(t *testing.T) {
	assert.True(t, cts.received)
	cts.received = false
}

// Close terminates the TestServer.
func (cts *TestServer) Close() {
	cts.server.Close()
}

// GenericMessage is a generic struct that can be used in tests.
type GenericMessage struct {
	Value string `json:"value"`
}