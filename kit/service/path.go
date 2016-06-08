package service

import (
	"golang.org/x/net/context"
	"net/http"
)

const (
	ctxLabelRequestPath = "requestPath"
)

// RequestPathExtractor is a go-kit before interceptor that puts the request path in the request context.
func RequestPathExtractor(ctx context.Context, r *http.Request) context.Context {
	return ctxWithRequestPath(ctx, r.URL.EscapedPath())
}

func ctxWithRequestPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, ctxLabelRequestPath, path)
}

func ctxRequestPath(ctx context.Context) string {
	return EnsureString(ctx, ctxLabelRequestPath)
}
