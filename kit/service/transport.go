package service

import (
	"encoding/json"
	"github.com/ConnectCorp/go-kit/kit/utils"
	"github.com/go-kit/kit/endpoint"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/dogstatsd"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/tylerb/graceful"
	"golang.org/x/net/context"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"net/http"
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
)

const (
	contentTypeHeaderName          = "Content-Type"
	jsonContentTypeHeaderValue     = "application/json"
	defaultShutdownLameDuckTimeout = 30 * time.Second
	healthRoutePath                = "/health"
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
func NewRouter(
	svcName, prefix string,
	rootLogger kitlog.Logger,
	tokenVerifier utils.TokenVerifier,
	dogstatsdEmitter *dogstatsd.Emitter,
	healthChecker HealthChecker) *Router {

	mux := mux.NewRouter()

	mux.Methods("GET").Path(healthRoutePath).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := healthChecker.CheckHealth(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	return &Router{
		rootCtx:         context.Background(),
		metricsReporter: NewMetricsReporter(commonMetricsNamespace, svcName, dogstatsdEmitter),
		transportLogger: NewTransportLogger(rootLogger, "REST"),
		tokenVerifier:   tokenVerifier,
		mux:             mux.PathPrefix(prefix).Subrouter(),
	}
}

// MountRoute mounts a Route on the Router.
func (r *Router) MountRoute(route Route) *Router {
	r.mux.Methods(route.GetMethod()).Path(route.GetPath()).Handler(kithttp.NewServer(
		r.rootCtx,
		r.chainMiddlewares(route.Endpoint,
			r.getTokenMiddleware(route.IsAuthenticated()),
			NewWireMiddleware(),
			NewMetricsMiddleware(r.metricsReporter),
			NewLoggingMiddleware(r.transportLogger)),
		route.Decoder,
		route.Encoder,
		kithttp.ServerBefore(WireExtractor, TokenExtractor, RequestPathExtractor, TraceIDExtractor),
		kithttp.ServerErrorEncoder(route.ErrorEncoder),
		kithttp.ServerAfter(TraceIDSetter)))

	return r
}

func (r *Router) getTokenMiddleware(authenticated bool) endpoint.Middleware {
	if authenticated {
		return NewTokenMiddleware(r.tokenVerifier)
	}
	return NewNoTokenMiddleware()
}

func (r *Router) chainMiddlewares(endpoint endpoint.Endpoint, middlewares ...endpoint.Middleware) endpoint.Endpoint {
	for _, middleware := range middlewares {
		endpoint = middleware(endpoint)
	}
	return endpoint
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
	method string
	path   string
}

// NewMethodAndPathMixin initializes a new MethodAndPathMixin.
func NewMethodAndPathMixin(method, path string) MethodAndPathMixin {
	return MethodAndPathMixin{method: method, path: path}
}

// GetMethod implements the Route interface.
func (m *MethodAndPathMixin) GetMethod() string {
	return m.method
}

// GetPath implements the Route interface.
func (m *MethodAndPathMixin) GetPath() string {
	return m.path
}

// AuthenticationMixin is a mixin implementing the Authentication interface.
type AuthenticationMixin struct {
	authenticated bool
}

// NewRequireAuthenticationMixin initializes a new AuthenticationMixin that requires authentication.
func NewRequireAuthenticationMixin() AuthenticationMixin {
	return AuthenticationMixin{authenticated: true}
}

// NewRejectAuthenticationMixin initializes a new AuthenticationMixin the rejects authentication.
func NewRejectAuthenticationMixin() AuthenticationMixin {
	return AuthenticationMixin{authenticated: false}
}

// IsAuthenticated implements the Authenticated interface.
func (a *AuthenticationMixin) IsAuthenticated() bool {
	return a.authenticated
}

// JSONEncoderMixin is a mixin implementing part of the Route interface.
type JSONEncoderMixin struct {
	// Intentionally empty.
}

// Encoder implements the Route interface.
func (*JSONEncoderMixin) Encoder(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
	w.Header().Add(contentTypeHeaderName, jsonContentTypeHeaderValue)
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
	w.Header().Add(contentTypeHeaderName, jsonContentTypeHeaderValue)
	w.WriteHeader(ErrorToStatusCode(err))
	_ = json.NewEncoder(w).Encode(&ErrorResponse{Error: err.Error()}) // Ignores an encoding error.
}

// JSONDecoderMixin is a mixin implementing part of the Route interface.
type JSONDecoderMixin struct {
	requestType reflect.Type
}

// MustNewJSONDecoderMixin initializes a new JSONDecoderMixin.
func MustNewJSONDecoderMixin(requestType interface{}) JSONDecoderMixin {
	t := reflect.TypeOf(requestType)
	if t.Kind() != reflect.Struct {
		panic(xerror.New("requestType must have kind = struct.", requestType))
	}
	return JSONDecoderMixin{requestType: t}
}

// Decoder implements the Route interface.
func (d *JSONDecoderMixin) Decoder(ctx context.Context, r *http.Request) (interface{}, error) {
	parsedBody := reflect.New(d.requestType).Interface()
	if _, err := utils.NewInboundRequest(r, parsedBody); err != nil {
		return nil, xerror.Wrap(err, ErrorBadRequest)
	}
	return parsedBody, nil
}
