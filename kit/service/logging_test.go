package service

import (
	"bytes"
	"encoding/json"
	"github.com/ConnectCorp/go-kit/kit/utils"
	"github.com/ConnectCorp/go-kit/kit/test"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"testing"
)

func TestLogging(t *testing.T) {
	w := bytes.NewBufferString("")

	ctx := context.Background()
	ctx = ctxWithTraceID(ctx, "trace-id")
	ctx = ctxWithRequestPath(ctx, "path")
	ctx = ctxWithClientType(ctx, "client-type")
	ctx = ctxWithClientVersion(ctx, "client-version")

	logger := utils.NewFormattedJSONLogger(w)
	loggingMiddleware := NewLoggingMiddleware(logger)
	loggingFunc := loggingMiddleware(test.TerminationMiddleware)
	_, err := loggingFunc(ctx, test.MustNewRequest())
	assert.Equal(t, "terminated", err.Error())

	parsedLogEntry := make(map[string]interface{})
	assert.Nil(t, json.Unmarshal(w.Bytes(), &parsedLogEntry))
	assert.Equal(t, "path", parsedLogEntry[actionKey].(string))
	assert.NotNil(t, parsedLogEntry[durationKey])
	assert.Equal(t, "trace-id", parsedLogEntry[ctxLabelTraceID])
	assert.Equal(t, "client-type", parsedLogEntry[ctxLabelClientType])
	assert.Equal(t, "client-version", parsedLogEntry[ctxLabelClientVersion])
	assert.NotNil(t, parsedLogEntry[requestKey])
	assert.NotNil(t, parsedLogEntry[errorKey])
}
