package service

import (
	"github.com/ConnectCorp/go-kit/kit/utils"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
	"golang.org/x/net/context"
	"io"
	"strings"
	"time"
)

const (
	actionKey   = "action"
	durationKey = "durationUs"
	requestKey  = "req"
	errorKey    = "err"
)

// NewRootLogger creates a root logger and configures the standard go log library.
func NewRootLogger(w io.Writer) kitlog.Logger {
	rootLogger := utils.NewFormattedJSONLogger(w)
	rootLogger = kitlog.NewContext(rootLogger).With("ts", utils.LogTimeRFC3339Nano)
	return rootLogger
}

// NewTransportLogger attaches the transport tag to the root logger.
func NewTransportLogger(rootLogger kitlog.Logger, transport string) kitlog.Logger {
	return kitlog.NewContext(rootLogger).With("transport", transport)
}

// NewBackgroundLogger attaches the background tag to the root logger.
func NewBackgroundLogger(rootLogger kitlog.Logger) kitlog.Logger {
	return kitlog.NewContext(rootLogger).With("background", true)
}

// NewLoggingMiddleware creates a new standard logging middleware for a Go microservice.
func NewLoggingMiddleware(logger kitlog.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (resp interface{}, err error) {
			defer func(startTime time.Time) {
				logRequest(logger, ctx, startTime, request, err)
			}(time.Now())
			return next(ctx, request)
		}
	}
}

func logRequest(logger kitlog.Logger, ctx context.Context, startTime time.Time, req interface{}, err error) {
	if err == nil {
		logger.Log(
			actionKey, ctxRequestPath(ctx),
			durationKey, durationUs(startTime),
			ctxLabelTraceID, CtxTraceID(ctx),
			ctxLabelClientType, ctxClientType(ctx),
			ctxLabelClientVersion, ctxClientVersion(ctx))
		return
	}
	logger.Log(
		actionKey, ctxRequestPath(ctx),
		durationKey, durationUs(startTime),
		ctxLabelTraceID, CtxTraceID(ctx),
		ctxLabelClientType, ctxClientType(ctx),
		ctxLabelClientVersion, ctxClientVersion(ctx),
		requestKey, strings.Split(spew.Sdump(req), "\n"),
		errorKey, err)
}

func durationUs(startTime time.Time) int64 {
	return time.Since(startTime).Nanoseconds() / 1000
}
