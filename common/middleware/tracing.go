package middleware

import (
	"github.com/ConnectCorp/go-kit/common/utils"
	"golang.org/x/net/context"
	"net/http"
)

const (
	traceIDLen      = 15
	traceIDHeader   = "X-Connect-Trace-ID"
	ctxLabelTraceID = "traceId"
)

func ctxWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ctxLabelTraceID, traceID)
}

func ctxTraceID(ctx context.Context) string {
	v := ctx.Value(ctxLabelTraceID)
	if v == nil {
		return ""
	}
	return v.(string)
}

func makeTraceID() string {
	return utils.GenRandomString(traceIDLen)
}

func TraceIDExtractor(ctx context.Context, r *http.Request) context.Context {
	traceID := r.Header.Get(traceIDHeader)

	if traceID != "" {
		ctx = ctxWithTraceID(ctx, traceID)
	} else {
		ctx = ctxWithTraceID(ctx, makeTraceID())
	}

	return ctx
}

func TraceIDSetter(ctx context.Context, w http.ResponseWriter) {
	w.Header().Set(traceIDHeader, ctxTraceID(ctx))
}
