package utils

import (
	"github.com/newrelic/go-agent/api"
	"github.com/newrelic/go-agent/api/datastore"
	"net/http"
)

// NoopNewrelicTransaction implements a noop newrelic.Transaction.
type NoopNewrelicTransaction struct {
	http.ResponseWriter
}

// End implements newrelic.Transaction.
func (*NoopNewrelicTransaction) End() error {
	return nil
}

// Ignore implements newrelic.Transaction.
func (*NoopNewrelicTransaction) Ignore() error {
	return nil
}

// SetName implements newrelic.Transaction.
func (*NoopNewrelicTransaction) SetName(name string) error {
	return nil
}

// NoticeError implements newrelic.Transaction.
func (*NoopNewrelicTransaction) NoticeError(err error) error {
	return nil
}

// AddAttribute implements newrelic.Transaction.
func (*NoopNewrelicTransaction) AddAttribute(key string, value interface{}) error {
	return nil
}

// StartSegment implements api.SegmentTracer.
func (*NoopNewrelicTransaction) StartSegment() api.Token {
	return 0
}

// EndSegment implements api.SegmentTracer.
func (*NoopNewrelicTransaction) EndSegment(token api.Token, name string) {
	// Noop.
}

// EndExternal implements api.SegmentTracer.
func (*NoopNewrelicTransaction) EndExternal(token api.Token, url string) {
	// Noop.
}

// EndDatastore implements api.SegmentTracer.
func (*NoopNewrelicTransaction) EndDatastore(api.Token, datastore.Segment) {
	// Noop.
}

// PrepareRequest implements api.SegmentTracer.
func (*NoopNewrelicTransaction) PrepareRequest(token api.Token, request *http.Request) {
	// Noop.
}

// EndRequest implements api.SegmentTracer.
func (*NoopNewrelicTransaction) EndRequest(token api.Token, request *http.Request, response *http.Response) {
	// Noop.
}

// NoopNewrelicApplication implements a noop newrelic.Application.
type NoopNewrelicApplication struct {
	// Intentionally empty.
}

// StartTransaction implements newrelic.Application.
func (*NoopNewrelicApplication) StartTransaction(name string, w http.ResponseWriter, r *http.Request) api.Transaction {
	return &NoopNewrelicTransaction{ResponseWriter: w}
}

// RecordCustomEvent implements newrelic.Application.
func (*NoopNewrelicApplication) RecordCustomEvent(eventType string, params map[string]interface{}) error {
	return nil
}
