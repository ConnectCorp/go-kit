package service

import (
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"net/http"
)

const (
	// ErrorMissingClientTypeHeader is returned when the X-Connect-Client-Type header is missing.
	ErrorMissingClientTypeHeader = "missing client type header"
	// ErrorMissingClientVersionHeader is returned when the X-Connect-Client-Version header is missing.
	ErrorMissingClientVersionHeader = "missing client version header"
)

const (
	clientTypeHeader      = "X-Connect-Client-Type"
	clientVersionHeader   = "X-Connect-Client-Version"
	ctxLabelClientType    = "clientType"
	ctxLabelClientVersion = "clientVersion"
)

// WireExtractor is a go-kit before handler that extracts common Connect headers into the request context.
func WireExtractor(ctx context.Context, r *http.Request) context.Context {
	clientType := r.Header.Get(clientTypeHeader)
	clientVersion := r.Header.Get(clientVersionHeader)

	if clientType != "" {
		ctx = ctxWithClientType(ctx, clientType)
	}
	if clientVersion != "" {
		ctx = ctxWithClientVersion(ctx, clientVersion)
	}

	return ctx
}

func ctxWithClientType(ctx context.Context, clientType string) context.Context {
	return context.WithValue(ctx, ctxLabelClientType, clientType)
}

func ctxClientType(ctx context.Context) string {
	return EnsureString(ctx, ctxLabelClientType)
}

func ctxWithClientVersion(ctx context.Context, clientVersion string) context.Context {
	return context.WithValue(ctx, ctxLabelClientVersion, clientVersion)
}

func ctxClientVersion(ctx context.Context) string {
	return EnsureString(ctx, ctxLabelClientVersion)
}

func checkCtx(ctx context.Context) error {
	if ctxClientType(ctx) == "" {
		return xerror.Wrap(xerror.New(ErrorMissingClientTypeHeader), ErrorBadRequest, ctx)
	}
	if ctxClientVersion(ctx) == "" {
		return xerror.Wrap(xerror.New(ErrorMissingClientVersionHeader), ErrorBadRequest, ctx)
	}
	return nil
}

// NewWireMiddleware creates a new standard wire middleware for a Go microservice.
func NewWireMiddleware() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (resp interface{}, err error) {
			if err := checkCtx(ctx); err != nil {
				return nil, err
			}
			return next(ctx, request)
		}
	}
}
