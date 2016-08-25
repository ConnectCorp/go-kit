package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ConnectCorp/go-kit/kit/test"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONDecoderMixin(t *testing.T) {
	decoder := MustNewJSONDecoderMixin(test.GenericMessage{})
	req, err := http.NewRequest("POST", "http://url", bytes.NewBufferString(`{ "value": "some-value" }`))
	assert.Nil(t, err)
	parsedReq, err := decoder.Decoder(context.Background(), req)
	assert.Nil(t, err)
	assert.Equal(t, &test.GenericMessage{"some-value"}, parsedReq)
}

func TestJSONEncoderMixin(t *testing.T) {
	recorder := httptest.NewRecorder()
	resp := &test.GenericMessage{"some-value"}
	assert.Nil(t, (&JSONEncoderMixin{}).Encoder(context.Background(), recorder, resp))
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, `{"data":{"value":"some-value"}}`+"\n", recorder.Body.String())
}

func TestJSONErrorEncoderMixin(t *testing.T) {
	recorder := httptest.NewRecorder()
	err := xerror.New(ErrorNotFound)
	(&JSONErrorEncoderMixin{}).ErrorEncoder(context.Background(), err, recorder)
	response := &ErrorResponse{}
	assert.Nil(t, json.Unmarshal(recorder.Body.Bytes(), response))
	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Equal(t, &ErrorResponse{Error: err.Error()}, response)
}

func TestAdvancedRouteMixin(t *testing.T) {
	_ = AdvancedRoute(NewAdvancedRouteMixin(true))
	assert.True(t, NewAdvancedRouteMixin(true).EnableWireMiddleware())
	assert.False(t, NewAdvancedRouteMixin(false).EnableWireMiddleware())
}

func TestErrorToStatusCode(t *testing.T) {
	assert.Equal(t, http.StatusBadRequest, ErrorToStatusCode(xerror.New(ErrorBadRequest)))
	assert.Equal(t, http.StatusUnauthorized, ErrorToStatusCode(xerror.New(ErrorUnauthorized)))
	assert.Equal(t, http.StatusForbidden, ErrorToStatusCode(xerror.New(ErrorForbidden)))
	assert.Equal(t, http.StatusNotFound, ErrorToStatusCode(xerror.New(ErrorNotFound)))
	assert.Equal(t, http.StatusInternalServerError, ErrorToStatusCode(xerror.New(ErrorUnexpected)))
	assert.Equal(t, http.StatusInternalServerError, ErrorToStatusCode(xerror.New("some-error")))
	assert.Equal(t, http.StatusInternalServerError, ErrorToStatusCode(fmt.Errorf("some-error")))
}
