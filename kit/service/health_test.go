package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRoutingHealthChecker(t *testing.T) {
	var h HealthChecker
	h = NewRoutingHealthChecker()
	assert.Nil(t, h.CheckHealth())
}

func TestCompoundHealthChecker(t *testing.T) {
	var h HealthChecker
	h = NewCompoundHealthChecker(NewRoutingHealthChecker())
	assert.Nil(t, h.CheckHealth())
}
