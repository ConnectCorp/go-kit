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
	"os"
	"strings"
	"testing"
)

type testRoute struct {
	AuthenticationMixin
	MethodAndPathMixin
	JSONEncoderMixin
	JSONErrorEncoderMixin
	AdvancedRouteMixin

	enableCors bool
}

type testRouteData struct{}

func newTestRoute(flag bool) *testRoute {
	return &testRoute{
		AuthenticationMixin: NewRequireAuthenticationMixin(),
		MethodAndPathMixin:  NewMethodAndPathMixin("GET", "/test"),
		AdvancedRouteMixin:  NewAdvancedRouteMixin(false, flag),
		enableCors:          flag,
	}
}

func (g *testRoute) CORSEnabled() bool { return g.enableCors }

func (g *testRoute) Endpoint(ctx context.Context, request interface{}) (interface{}, error) {
	return nil, nil
}

func (g *testRoute) Decoder(_ context.Context, r *http.Request) (interface{}, error) { return nil, nil }

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
	mixin := NewAdvancedRouteMixin(true, true)
	assert.True(t, (&mixin).EnableWireMiddleware())
	assert.True(t, (&mixin).EnableCORSMiddleware())
	mixin = NewAdvancedRouteMixin(false, false)
	assert.False(t, (&mixin).EnableWireMiddleware())
	assert.False(t, (&mixin).EnableCORSMiddleware())

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

func TestRouteWithoutCORS(t *testing.T) {
	rootLogger := NewRootLogger(os.Stdout)
	router := NewRouter("test", "/v1", rootLogger, nil, nil, nil, nil)
	router.MountRoute(newTestRoute(false))

	ts := httptest.NewServer(router.GetMux())
	defer ts.Close()

	req, _ := http.NewRequest("OPTIONS", ts.URL+"/v1/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "PUT")
	req.Header.Set("Access-Control-Request-headers", "Bad-Header-Type")
	res, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, 404)
}

func TestRouteWithCORS(t *testing.T) {
	rootLogger := NewRootLogger(os.Stdout)
	router := NewRouter("test", "/v1", rootLogger, nil, nil, nil, nil)
	router.MountRoute(newTestRoute(true))

	ts := httptest.NewServer(router.GetMux())
	defer ts.Close()

	req, _ := http.NewRequest("OPTIONS", ts.URL+"/v1/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "PUT")
	req.Header.Set("Access-Control-Request-headers", "Bad-Header-Type")
	res, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, 403)

	req, _ = http.NewRequest("OPTIONS", ts.URL+"/v1/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "PUT")
	req.Header.Set("Access-Control-Request-headers", "X-Connect-Client-Type")
	res, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, 200)
}

func TestCORSBadHeaders(t *testing.T) {
	var handler http.Handler
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "http://example.com/foo", nil)
	req.Header.Add("Origin", "http://example.com")
	req.Header.Add("Access-Control-Request-Method", "PUT")
	req.Header.Add("Access-Control-Request-headers", "Bad-Header-Type")
	handler = corsMiddleware(handler)
	handler.ServeHTTP(res, req)
	assert.Equal(t, res.Code, 403)
	assertHeaders(t, res.Header(), map[string]string{
		"Origin":                       "",
		"Access-Control-Allow-Methods": "",
		"Access-Control-Allow-Headers": "",
	})
}

func TestCORSBadMethod(t *testing.T) {
	var handler http.Handler
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "http://example.com/foo", nil)
	req.Header.Add("Origin", "http://example.com")
	req.Header.Add("Access-Control-Request-Method", "OPTIONS")
	req.Header.Add("Access-Control-Request-headers", "X-Connect-Client-Type")
	handler = corsMiddleware(handler)
	handler.ServeHTTP(res, req)
	assert.Equal(t, res.Code, 405)
	assertHeaders(t, res.Header(), map[string]string{
		"Origin":                       "",
		"Access-Control-Allow-Methods": "",
		"Access-Control-Allow-Headers": "",
	})
}

func TestCORSCorrectHeaders(t *testing.T) {
	var handler http.Handler
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "http://example.com/foo", nil)
	req.Header.Add("Origin", "http://example.com")
	req.Header.Add("Access-Control-Request-Method", "PUT")
	req.Header.Add("Access-Control-Request-headers", "X-Connect-Client-Type")
	handler = corsMiddleware(handler)
	handler.ServeHTTP(res, req)
	assert.Equal(t, res.Code, 200)
	assertHeaders(t, res.Header(), map[string]string{
		"Origin":                       "",
		"Access-Control-Allow-Methods": "PUT",
		"Access-Control-Allow-Headers": "X-Connect-Client-Type",
	})
}

func assertHeaders(t *testing.T, resHeaders http.Header, reqHeaders map[string]string) {
	for name, value := range reqHeaders {
		if actual := strings.Join(resHeaders[name], ", "); actual != value {
			t.Errorf("Invalid header `%s', wanted `%s', got `%s'", name, value, actual)
		}
	}
}
