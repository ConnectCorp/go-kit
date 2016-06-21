package utils

import (
	"encoding/json"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"io/ioutil"
	"net/http"
)

const (
	// ErrorCannotReadBody is returned when an unexpected error occurs while reading the request/response body.
	ErrorCannotReadBody = "cannot read body"
	// ErrorBodyMustBeEmpty is returned when an unnecessary request body is provided.
	ErrorBodyMustBeEmpty = "body must be empty"
	// ErrorCannotParseJSON is returned when the body of the request cannot be parsed as JSON.
	ErrorCannotParseJSON = "cannot parse json"
	// ErrorClient is returned when the HTTP client failed to complete a request.
	ErrorClient = "client error"
)

// InboundRequest is a thin layer on top of an HTTP request.
type InboundRequest struct {
	request    *http.Request
	cachedBody []byte
}

// NewInboundRequest initializes a new InboundRequest.
func NewInboundRequest(req *http.Request, parsedBody interface{}) (*InboundRequest, error) {
	ir := &InboundRequest{request: req}

	body, err := ioutil.ReadAll(ir.request.Body) // This is auto-closed by the handler.
	if err != nil {
		return nil, xerror.Wrap(err, ErrorCannotReadBody)
	}
	ir.cachedBody = body

	if parsedBody == nil && len(ir.cachedBody) > 0 {
		return nil, xerror.New(ErrorBodyMustBeEmpty)
	}

	if parsedBody != nil {
		if err := json.Unmarshal(ir.cachedBody, parsedBody); err != nil {
			return nil, xerror.Wrap(err, ErrorCannotParseJSON)
		}
	}

	return ir, nil
}

// GetRequest returns the underlying HTTP request.
func (ir *InboundRequest) GetRequest() *http.Request {
	return ir.request
}

// GetCachedBody returns the body of the request.
func (ir *InboundRequest) GetCachedBody() []byte {
	return ir.cachedBody
}

// InboundResponse is a thing layer on top of an HTTP response.
type InboundResponse struct {
	response   *http.Response
	cachedBody []byte
}

// NewInboundResponse initializes a new InboundResponse.
func NewInboundResponse(resp *http.Response, err error) (*InboundResponse, error) {
	if err != nil {
		return nil, xerror.Wrap(err, ErrorClient)
	}

	ir := &InboundResponse{response: resp}

	body, err := ioutil.ReadAll(ir.response.Body)
	ir.response.Body.Close() // Swallow any close error, there's nothing we can do.
	if err != nil {
		return nil, xerror.Wrap(err, ErrorCannotReadBody)
	}
	ir.cachedBody = body

	return ir, nil
}

// GetResponse returns the underlying HTTP response.
func (ir *InboundResponse) GetResponse() *http.Response {
	return ir.response
}

// GetCachedBody returns the response body.
func (ir *InboundResponse) GetCachedBody() []byte {
	return ir.cachedBody
}

// IsSuccessful returns true if the status code is successful (2xx).
func (ir *InboundResponse) IsSuccessful() bool {
	return ir.response.StatusCode >= 200 && ir.response.StatusCode <= 299
}

// ParseJSON parses the response body into the given interface.
func (ir *InboundResponse) ParseJSON(parsedBody interface{}) error {
	if err := json.Unmarshal(ir.cachedBody, parsedBody); err != nil {
		return xerror.Wrap(err, ErrorCannotParseJSON)
	}
	return nil
}
