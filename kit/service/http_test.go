package service

import (
	"testing"
	"github.com/ConnectCorp/go-kit/kit/test"
	"net/http"
	"github.com/stretchr/testify/assert"
)

func TestProdHTTPClient(t *testing.T) {
	ts := test.NewTestServer()
	defer ts.Close()

	client := MakeProdHTTPClient()

	requestCount := 0
	ts.SetResponder(func(w http.ResponseWriter, r *http.Request) {
		if requestCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		requestCount++
	})
	_, err := client.Get(ts.URL("/test1"))
	assert.Nil(t, err)
	assert.Equal(t, 4, requestCount)

	requestCount = 0
	ts.SetResponder(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusInternalServerError)
	})
	resp, err := client.Get(ts.URL("/test1"))
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, 4, requestCount)
}
