package service

import (
	"github.com/ConnectCorp/go-kit/kit/utils"
	"golang.org/x/net/context"
	"net/http"
)

const (
	traceIDLen      = 15
	traceIDHeader   = "X-Connect-Trace-ID"
	ctxLabelTraceID = "traceId"
)

// TraceIDExtractor is a go-kit before handler that extracts a trace ID from HTTP headers, or creates a new one.
func TraceIDExtractor(ctx context.Context, r *http.Request) context.Context {
	traceID := r.Header.Get(traceIDHeader)
	if traceID != "" {
		return ctxWithTraceID(ctx, traceID)
	}
	return ctxWithTraceID(ctx, utils.GenRandomString(traceIDLen))
}

// TraceIDSetter is a go-kit after handler that sets the context trace ID into a HTTP header.
func TraceIDSetter(ctx context.Context, w http.ResponseWriter) {
	w.Header().Set(traceIDHeader, ctxTraceID(ctx))
}

func ctxWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ctxLabelTraceID, traceID)
}

func ctxTraceID(ctx context.Context) string {
	return EnsureString(ctx, ctxLabelTraceID)
}
