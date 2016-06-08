package service

import (
	"encoding/json"
	kithttp "github.com/go-kit/kit/transport/http"
	"golang.org/x/net/context"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"net/http"
	"reflect"
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
	defaultContentTypeHeaderName  = "Content-Type"
	defaultContentTypeHeaderValue = "application/json"
)

// Response is the standard API successful response container Go microservices.
type Response struct {
	Data interface{} `json:"data,omitempty"`
}

// ErrorResponse is the standard API error response for Go microservices.
type ErrorResponse struct {
	Error string `json:"error,omitempty"`
}

// MakePOSTRequestDecoder generates a go-kit request decoder that parses the request body as JSON.
func MakePOSTRequestDecoder(requestType reflect.Type) kithttp.DecodeRequestFunc {
	if requestType.Kind() != reflect.Struct {
		panic(xerror.New("requestType must have kind = struct.", requestType))
	}
	return func(ctx context.Context, r *http.Request) (interface{}, error) {
		req := reflect.New(requestType).Interface()
		err := json.NewDecoder(r.Body).Decode(req)
		defer r.Body.Close()
		if err != nil {
			return nil, xerror.Wrap(xerror.Wrap(err, ErrorCannotDecode, r), ErrorBadRequest)
		}
		return req, nil
	}
}

// EncodeResponseJSON is a go-kit transport encoder for successful responses.
func EncodeResponseJSON(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
	w.Header().Add(defaultContentTypeHeaderName, defaultContentTypeHeaderValue)
	return json.NewEncoder(w).Encode(&Response{resp})
}

// EncodeErrorJSON is a go-kit transport encoder for errors.
func EncodeErrorJSON(_ context.Context, err error, w http.ResponseWriter) {
	if kitErr, ok := err.(kithttp.Error); ok {
		err = kitErr.Err
	}
	w.Header().Add(defaultContentTypeHeaderName, defaultContentTypeHeaderValue)
	w.WriteHeader(ErrorToStatusCode(err))
	_ = json.NewEncoder(w).Encode(&ErrorResponse{Error: err.Error()}) // Ignores an encoding error.
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
