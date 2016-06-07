package service

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTraceIDExtractor(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost", nil)
	assert.Nil(t, err)
	assert.True(t, len(EnsureString(TraceIDExtractor(context.Background(), req), ctxLabelTraceID)) == traceIDLen)
	req.Header.Set(traceIDHeader, "trace-id")
	assert.True(t, EnsureString(TraceIDExtractor(context.Background(), req), ctxLabelTraceID) == "trace-id")
}

func TestTraceIDSetter(t *testing.T) {
	recorder := httptest.NewRecorder()
	TraceIDSetter(context.Background(), recorder)
	assert.Equal(t, "", recorder.HeaderMap.Get(traceIDHeader))
	recorder = httptest.NewRecorder()
	TraceIDSetter(ctxWithTraceID(context.Background(), "trace-id"), recorder)
	assert.Equal(t, "trace-id", recorder.HeaderMap.Get(traceIDHeader))
}
