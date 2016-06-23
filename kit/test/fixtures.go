package test

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"net/http"
	"net/http/httptest"
	"testing"
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

// TempServer is a test HTTP server that validates incoming requests.
type TempServer struct {
	server    *httptest.Server
	validator func(r *http.Request)
	responder func(w http.ResponseWriter, r *http.Request)
	received  bool
}

// NewTempServer initialized a new TestServer.
func NewTempServer() *TempServer {
	cts := &TempServer{}
	cts.server = httptest.NewServer(cts)
	return cts
}

// URL return a URL to the test server with the given path.
func (ts *TempServer) URL(path string) string {
	return ts.server.URL + path
}

// SetValidator sets the validator function used for validating a request.
func (ts *TempServer) SetValidator(validator func(r *http.Request)) {
	ts.validator = validator
}

// SetResponder sets the responder function used for returning a response.
func (ts *TempServer) SetResponder(responder func(w http.ResponseWriter, r *http.Request)) {
	ts.responder = responder
}

// ServeHTTP implements the http.Handler interface.
func (ts *TempServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ts.received = true
	if ts.validator != nil {
		ts.validator(r)
	}
	if ts.responder != nil {
		ts.responder(w, r)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

// AssertReceived asserts that a request has been received and resets the server.
func (ts *TempServer) AssertReceived(t *testing.T) {
	assert.True(t, ts.received)
	ts.received = false
}

// Close terminates the TestServer.
func (ts *TempServer) Close() {
	ts.server.Close()
}

// GenericMessage is a generic struct that can be used in tests.
type GenericMessage struct {
	Value string `json:"value"`
}
