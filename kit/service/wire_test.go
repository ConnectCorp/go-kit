package service

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"testing"
)

func TestWire(t *testing.T) {
	req := mustTestRequest()
	req.Header.Set(clientTypeHeader, "client-type")
	req.Header.Set(clientVersionHeader, "client-version")
	ctx := WireExtractor(context.Background(), req)
	assert.Equal(t, "client-type", ctxClientType(ctx))
	assert.Equal(t, "client-version", ctxClientVersion(ctx))

	wireMiddleware := NewWireMiddleware()
	wireFunc := wireMiddleware(testTerminationMiddleware)

	_, err := wireFunc(ctx, req)
	assert.Equal(t, "terminated", err.Error())

	_, err = wireFunc(context.Background(), req)
	assert.Equal(t, "bad request: missing client type header", err.Error())
}
