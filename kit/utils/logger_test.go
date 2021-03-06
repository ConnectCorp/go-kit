package utils

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testLogEntry struct {
	Value string `json:"value"`
}

var expectedLogString = `{"k1":"v1","k2":{"value":"test-value"}}` + "\n"

func TestLogger(t *testing.T) {
	w := bytes.NewBufferString("")
	l := NewFormattedJSONLogger(w)
	l.Log("k1", "v1", "k2", &testLogEntry{"test-value"})
	assert.Equal(t, expectedLogString, w.String())
}
