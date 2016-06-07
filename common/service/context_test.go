package service

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"testing"
)

func TestEnsureString(t *testing.T) {
	assert.Equal(t, "value", EnsureString(context.WithValue(context.Background(), "key", "value"), "key"))
	assert.Equal(t, "", EnsureString(context.Background(), "key"))
}

func TestEnsureInt64(t *testing.T) {
	assert.Equal(t, int64(1), EnsureInt64(context.WithValue(context.Background(), "key", int64(1)), "key"))
	assert.Equal(t, int64(0), EnsureInt64(context.Background(), "key"))
}
