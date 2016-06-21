package utils

import (
	"bytes"
	"github.com/ConnectCorp/go-kit/kit/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestInboundRequest(t *testing.T) {
	ts := test.NewTestServer()
	defer ts.Close()

	// No body, no body requested.
	ts.SetValidator(func(r *http.Request) {
		ir, err := NewInboundRequest(r, nil)
		assert.Nil(t, err)
		assert.Equal(t, "/test1", ir.GetRequest().URL.String())
		assert.Equal(t, 0, len(ir.GetCachedBody()))
	})
	_, err := http.Get(ts.URL("/test1"))
	assert.Nil(t, err)
	ts.AssertReceived(t)

	// No body, body requested.
	ts.SetValidator(func(r *http.Request) {
		_, err := NewInboundRequest(r, &test.GenericMessage{})
		assert.True(t, xerror.Is(err, ErrorCannotParseJSON))
	})
	_, err = http.Get(ts.URL("/test1"))
	assert.Nil(t, err)
	ts.AssertReceived(t)

	// Body, no body requested.
	ts.SetValidator(func(r *http.Request) {
		_, err := NewInboundRequest(r, nil)
		assert.True(t, xerror.Is(err, ErrorBodyMustBeEmpty))
	})
	_, err = http.Post(ts.URL("/test1"), "application/json", bytes.NewBufferString("{}"))
	assert.Nil(t, err)
	ts.AssertReceived(t)

	// Body, body requested. Also check the body is closed.
	var request *http.Request
	ts.SetValidator(func(r *http.Request) {
		request = r
		ir, err := NewInboundRequest(r, &test.GenericMessage{})
		assert.Nil(t, err)
		assert.Equal(t, []byte(`{ "value": "v1" }`), ir.GetCachedBody())
	})
	_, err = http.Post(ts.URL("/test1"), "application/json", bytes.NewBufferString(`{ "value": "v1" }`))
	assert.Nil(t, err)
	ts.AssertReceived(t)
	_, err = ioutil.ReadAll(request.Body)
	assert.Equal(t, "http: invalid Read on closed Body", err.Error())
}

func TestInboundResponse(t *testing.T) {
	ts := test.NewTestServer()
	defer ts.Close()

	// Successful, empty body.
	ir, err := NewInboundResponse(http.Get(ts.URL("/test1")))
	assert.Nil(t, err)
	assert.True(t, ir.IsSuccessful())
	assert.Equal(t, 0, len(ir.GetCachedBody()))
	ts.AssertReceived(t)

	// Succesful, body.
	ts.SetResponder(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{ "value": "v1" }`))
		assert.Nil(t, err)
	})
	ir, err = NewInboundResponse(http.Get(ts.URL("/test1")))
	assert.Nil(t, err)
	assert.True(t, ir.IsSuccessful())
	assert.Equal(t, []byte(`{ "value": "v1" }`), ir.GetCachedBody())
	gm := &test.GenericMessage{}
	assert.Nil(t, ir.ParseJSON(gm))
	assert.Equal(t, "v1", gm.Value)
	assert.True(t, xerror.Is(ir.ParseJSON(map[string]string{}), ErrorCannotParseJSON))
	ts.AssertReceived(t)
	_, err = ioutil.ReadAll(ir.GetResponse().Body)
	assert.Equal(t, "http: read on closed response body", err.Error())

	// Not successful.
	ts.SetResponder(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	ir, err = NewInboundResponse(http.Get(ts.URL("/test1")))
	assert.Nil(t, err)
	assert.False(t, ir.IsSuccessful())
	ts.AssertReceived(t)

	// Failed request.
	_, err = NewInboundResponse(http.Get("bad"))
	assert.True(t, xerror.Is(err, ErrorClient))
}
