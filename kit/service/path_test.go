package service

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"net/http"
	"testing"
)

func TestRequestPathExtractor(t *testing.T) {
	req, err := http.NewRequest("POST", "http://url/path", nil)
	assert.Nil(t, err)
	assert.Equal(t, "/path", ctxRequestPath(RequestPathExtractor(context.Background(), req)))
}
