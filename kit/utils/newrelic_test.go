package utils

import (
	"github.com/newrelic/go-agent"
	"testing"
)

func TestNoopNewrelicApplication(t *testing.T) {
	_ = newrelic.Application(&NoopNewrelicApplication{})
}

func TestNoopNewrelicTransaction(t *testing.T) {
	_ = newrelic.Transaction(&NoopNewrelicTransaction{})
}
