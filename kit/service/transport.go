package service

import (
	"encoding/json"
	"github.com/ConnectCorp/go-kit/kit/utils"
	kitlog "github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/tylerb/graceful"
	"golang.org/x/net/context"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"net/http"
	"os"
	"reflect"
	"time"
)

const (
	// ErrorBadRequest is returned when a request cannot be parsed or is otherwise incorrect.
	ErrorBadRequest = "bad request"
	// ErrorUnauthorized is returned when a required auth token is missing or invalid.
	ErrorUnauthorized = "unauthorized"
	// ErrorForbidden is returned when the auth token is valid but the user doesn't have sufficient permissions.
	ErrorForbidden = "forbidden"
	// ErrorNotFound is returned when the requested resource cannot be found.
	ErrorNotFound = "not found"
	// ErrorUnexpected is returned when no more specific error can be isolated.
	ErrorUnexpected = "unexpected"
	// ErrorCannotDecode is returned when decoding a request body fails.
	ErrorCannotDecode = "cannot decode request body"
)

const (
	defaultContentTypeHeaderName   = "Content-Type"
	defaultContentTypeHeaderValue  = "application/json"
	defaultShutdownLameDuckTimeout = 30 * time.Second
)

// Response is the standard API successful response container Go microservices.
type Response struct {
	Data interface{} `json:"data,omitempty"`
}

// ErrorResponse is the standard API error response for Go microservices.
type ErrorResponse struct {
	Error string `json:"error,omitempty"`
}

// ErrorToStatusCode converts an error to the corresponding HTTP status code.
func ErrorToStatusCode(err error) int {
	if xerror.Is(err, ErrorBadRequest) {
		return http.StatusBadRequest
	}
	if xerror.Is(err, ErrorUnauthorized) {
		return http.StatusUnauthorized
	}
	if xerror.Is(err, ErrorForbidden) {
		return http.StatusForbidden
	}
	if xerror.Is(err, ErrorNotFound) {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

// Router implements a router for Connect microservices.
type Router struct {
	prefix          string
	rootCtx         context.Context
	metricsReporter MetricsReporter
	transportLogger kitlog.Logger
	tokenVerifier   utils.TokenVerifier
	mux             *mux.Router
}

// NewRouter initializes a new Router.
func NewRouter(svcName, prefix string, tokenVerifier utils.TokenVerifier) *Router {
	return &Router{
		rootCtx:         context.Background(),
		metricsReporter: NewMetricsReporter(commonMetricsNamespace, svcName),
		transportLogger: NewTransportLogger(utils.NewFormattedJSONLogger(os.Stderr), "REST"),
		tokenVerifier:   tokenVerifier,
		mux:             mux.NewRouter().PathPrefix(prefix).Subrouter(),
	}
}

// MountRoute mounts a Route on the Router.
func (r *Router) MountRoute(route Route) *Router {
	endpoint := route.Endpoint

	if route.IsAuthenticated() {
		endpoint = NewTokenMiddleware(r.tokenVerifier)(endpoint)
	} else {
		endpoint = NewNoTokenMiddleware()(endpoint)
	}

	endpoint = NewWireMiddleware()(endpoint)
	endpoint = NewMetricsMiddleware(r.metricsReporter)(endpoint)
	endpoint = NewLoggingMiddleware(r.transportLogger)(endpoint)

	r.mux.Methods(route.GetMethod()).Path(route.GetPath()).Handler(kithttp.NewServer(
		r.rootCtx,
		endpoint,
		route.Decoder,
		route.Encoder,
		kithttp.ServerBefore(WireExtractor, TokenExtractor, RequestPathExtractor, TraceIDExtractor),
		kithttp.ServerErrorEncoder(route.ErrorEncoder),
		kithttp.ServerAfter(TraceIDSetter)))

	return r
}

// GetMux returns the underlying Gorilla *mux.Router, useful for testing or custom configuration.
func (r *Router) GetMux() *mux.Router {
	return r.mux
}

// Run exposes the Router on the given address spec. Blocks forever, or until a fatal error occurs.
func (r *Router) Run(addr string) {
	graceful.Run(addr, defaultShutdownLameDuckTimeout, r.mux)
}

// Route describes a route to an endpoint in a Router.
type Route interface {
	Authentication

	GetMethod() string
	GetPath() string
	Endpoint(ctx context.Context, request interface{}) (response interface{}, err error) // endpoint.Endpoint
	Decoder(context.Context, *http.Request) (request interface{}, err error)             // kithttp.DecodeRequestFunc
	Encoder(context.Context, http.ResponseWriter, interface{}) error                     // kithttp.EncodeResponseFunc
	ErrorEncoder(ctx context.Context, err error, w http.ResponseWriter)                  // kithttp.ErrorEncoder
}

// Authentication describes whether an endpoint should be authenticated.
type Authentication interface {
	IsAuthenticated() bool
}

// MethodAndPathMixin is a mixin implementing part of the Route interface.
type MethodAndPathMixin struct {
	Method string
	Path   string
}

// GetMethod implements the Route interface.
func (m *MethodAndPathMixin) GetMethod() string {
	return m.Method
}

// GetPath implements the Route interface.
func (m *MethodAndPathMixin) GetPath() string {
	return m.Path
}

// AuthenticationMixin is a mixin implementing the Authentication interface.
type AuthenticationMixin struct {
	Authenticated bool
}

// IsAuthenticated implements the Authenticated interface.
func (a *AuthenticationMixin) IsAuthenticated() bool {
	return a.Authenticated
}

// JSONEncoderMixin is a mixin implementing part of the Route interface.
type JSONEncoderMixin struct {
	// Intentionally empty.
}

// Encoder implements the Route interface.
func (*JSONEncoderMixin) Encoder(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
	w.Header().Add(defaultContentTypeHeaderName, defaultContentTypeHeaderValue)
	return json.NewEncoder(w).Encode(&Response{resp})
}

// JSONErrorEncoderMixin is a mixin implementing part of the Route interface.
type JSONErrorEncoderMixin struct {
	// Intentionally empty.
}

// ErrorEncoder implements the Route interface.
func (*JSONErrorEncoderMixin) ErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	if kitErr, ok := err.(kithttp.Error); ok {
		err = kitErr.Err
	}
	w.Header().Add(defaultContentTypeHeaderName, defaultContentTypeHeaderValue)
	w.WriteHeader(ErrorToStatusCode(err))
	_ = json.NewEncoder(w).Encode(&ErrorResponse{Error: err.Error()}) // Ignores an encoding error.
}

// JSONDecoderMixin is a mixin implementing part of the Route interface.
type JSONDecoderMixin struct {
	RequestType reflect.Type
}

// Decoder implements the Route interface.
func (d *JSONDecoderMixin) Decoder(ctx context.Context, r *http.Request) (interface{}, error) {
	req := reflect.New(d.RequestType).Interface()
	err := json.NewDecoder(r.Body).Decode(req)
	defer r.Body.Close()
	if err != nil {
		return nil, xerror.Wrap(xerror.Wrap(err, ErrorCannotDecode, r), ErrorBadRequest)
	}
	return req, nil
}
