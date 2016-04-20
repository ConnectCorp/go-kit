package middleware

import (
	"github.com/ConnectCorp/go-kit/common/utils"
	"golang.org/x/net/context"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
)

const (
	ErrorMissingToken        = "missing token"
	ErrorMustNotAuthenticate = "must not authenticate"
	ErrorUnauthorized        = "unauthorized"

	ctxLabelToken          = "token"
	ctxLabelAuthorizedSub  = "authorizedSub"
	ctxLabelAuthorizedRole = "authorizedRole"
)

type token struct {
	jwtPublicKey []byte
}

func (t *token) RequireToken(ctx context.Context) (context.Context, error) {
	token := CtxToken(ctx)
	if token == "" {
		return ctx, xerror.New(ErrorMissingToken, ctx)
	}
	return t.CheckToken(ctx)
}

func (t *token) CheckToken(ctx context.Context) (context.Context, error) {
	token := CtxToken(ctx)
	if token == "" {
		return CtxWithAuthorizedSub(ctx, 0), nil
	}
	userID, role, err := utils.VerifyToken(token, t.jwtPublicKey)
	if err != nil {
		return ctx, xerror.Wrap(err, ErrorUnauthorized)
	}
	return CtxWithAuthorizedRole(CtxWithAuthorizedSub(ctx, userID), role), nil
}

func (t *token) ForbidToken(ctx context.Context) (context.Context, error) {
	token := CtxToken(ctx)
	if token != "" {
		return ctx, xerror.New(ErrorMustNotAuthenticate, ctx)
	}
	return CtxWithAuthorizedSub(ctx, 0), nil
}

func CtxWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, ctxLabelToken, token)
}

func CtxToken(ctx context.Context) string {
	v := ctx.Value(ctxLabelToken)
	if v == nil {
		return ""
	}
	return v.(string)
}

func CtxWithAuthorizedSub(ctx context.Context, authorizedUserID int64) context.Context {
	return context.WithValue(ctx, ctxLabelAuthorizedSub, authorizedUserID)
}

func CtxAuthorizedSub(ctx context.Context) int64 {
	v := ctx.Value(ctxLabelAuthorizedSub)
	if v == nil {
		return 0
	}
	return v.(int64)
}

func CtxWithAuthorizedRole(ctx context.Context, authorizedRole string) context.Context {
	return context.WithValue(ctx, ctxLabelAuthorizedRole, authorizedRole)
}

func CtxAuthorizedRole(ctx context.Context) string {
	v := ctx.Value(ctxLabelAuthorizedRole)
	if v == nil {
		return ""
	}
	return v.(string)
}
