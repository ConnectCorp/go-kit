package service

import (
	"github.com/ConnectCorp/go-kit/kit/utils"
	"github.com/go-kit/kit/endpoint"
	"golang.org/x/net/context"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"net/http"
	"strings"
)

const (
	// ErrorMissingToken is returned when a required auth token is missing.
	ErrorMissingToken = "missing token"
	// ErrorMustNotAuthenticate is returned when a token is provided to a request that does not need it.
	ErrorMustNotAuthenticate = "must not authenticate"
)

const (
	authorizationHeader    = "Authorization"
	ctxLabelToken          = "token"
	ctxLabelAuthorizedSub  = "authorizedSub"
	ctxLabelAuthorizedRole = "authorizedRole"
)

func ctxWithAuthorizedRole(ctx context.Context, authorizedRole string) context.Context {
	return context.WithValue(ctx, ctxLabelAuthorizedRole, authorizedRole)
}

// CtxAuthorizedRole extracts the verified role stored in the request context.
func CtxAuthorizedRole(ctx context.Context) string {
	return EnsureString(ctx, ctxLabelAuthorizedRole)
}

func ctxWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, ctxLabelToken, token)
}

func ctxToken(ctx context.Context) string {
	return EnsureString(ctx, ctxLabelToken)
}

func ctxWithAuthorizedSub(ctx context.Context, authorizedUserID int64) context.Context {
	return context.WithValue(ctx, ctxLabelAuthorizedSub, authorizedUserID)
}

// CtxAuthorizedSub extracts the verified sub stored in the request context.
func CtxAuthorizedSub(ctx context.Context) int64 {
	return EnsureInt64(ctx, ctxLabelAuthorizedSub)
}

// TokenExtractor is a go-kit before handler that extracts a token from the Authorization header into the context.
func TokenExtractor(ctx context.Context, r *http.Request) context.Context {
	token := r.Header.Get(authorizationHeader)
	if token != "" {
		return ctxWithToken(ctx, token)
	}
	return ctx
}

// NewTokenMiddleware requires a valid token for the request, attaches verified role and sub in the context.
func NewTokenMiddleware(tokenVerifier utils.TokenVerifier) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (resp interface{}, err error) {
			if ctx, err = requireToken(ctx, tokenVerifier); err != nil {
				return nil, err
			}
			return next(ctx, request)
		}
	}
}

// NewNoTokenMiddleware requires that no token is attached to the request.
func NewNoTokenMiddleware() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (resp interface{}, err error) {
			if ctx, err = forbidToken(ctx); err != nil {
				return nil, err
			}
			return next(ctx, request)
		}
	}
}

func requireToken(ctx context.Context, tokenVerifier utils.TokenVerifier) (context.Context, error) {
	token := getToken(ctx)
	if token == "" {
		return ctx, xerror.Wrap(xerror.New(ErrorMissingToken), ErrorUnauthorized, ctx)
	}
	return checkToken(ctx, tokenVerifier)
}

func checkToken(ctx context.Context, tokenVerifier utils.TokenVerifier) (context.Context, error) {
	token := getToken(ctx)
	if token == "" {
		return ctxWithAuthorizedSub(ctx, 0), nil
	}
	userID, role, err := tokenVerifier.VerifyToken(token)
	if err != nil {
		return ctx, xerror.Wrap(err, ErrorUnauthorized)
	}
	return ctxWithAuthorizedRole(ctxWithAuthorizedSub(ctx, userID), role), nil
}

func forbidToken(ctx context.Context) (context.Context, error) {
	token := getToken(ctx)
	if token != "" {
		return ctx, xerror.Wrap(xerror.New(ErrorMustNotAuthenticate), ErrorBadRequest, ctx)
	}
	return ctxWithAuthorizedSub(ctx, 0), nil
}

func getToken(ctx context.Context) string {
	token := ctxToken(ctx)
	if token == "" {
		return token
	}
	return strings.TrimSpace(strings.TrimPrefix(token, "Bearer"))
}
