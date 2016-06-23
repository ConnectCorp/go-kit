package server

import (
	"github.com/ConnectCorp/go-kit/kit/service"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
)

const (
	publicSpec  = ":10000"
	privateSpec = ":10001"
)

// RunServer runs a server forever, until an error occurs.
func RunServer(router *service.Router) {
	txErrChan := make(chan error)
	http.Handle("/metrics", prometheus.Handler())
	go func() { router.Run(":10000") }()
	go func() { txErrChan <- http.ListenAndServe(":10001", nil) }()
	log.Printf("exit: %v\n", <-txErrChan)
}
