package service

import (
	"github.com/ConnectCorp/go-kit/kit/test"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestProdHTTPClient(t *testing.T) {
	ts := test.NewTempServer()
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
