package service

import (
	"golang.org/x/net/context"
)

const (
	ctxLabelRequestPath = "requestPath"
)

// EnsureString extracts a string from the context, returns "" if not found.
func EnsureString(ctx context.Context, label string) string {
	v := ctx.Value(label)
	if v == nil {
		return ""
	}
	return v.(string)
}

// EnsureInt64 extracts an int64 from the context, returns 0 if not found.
func EnsureInt64(ctx context.Context, label string) int64 {
	v := ctx.Value(label)
	if v == nil {
		return 0
	}
	return v.(int64)
}
