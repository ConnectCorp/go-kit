package service

import (
	"golang.org/x/net/context"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"net/http"
)

func mustTestRequest() *http.Request {
	req, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		panic(err)
	}
	return req
}

func testTerminationMiddleware(_ context.Context, _ interface{}) (interface{}, error) {
	return nil, xerror.New("terminated")
}
