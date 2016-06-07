package service

import (
	"golang.org/x/net/context"
	"testing"
	"time"
)

func TestSomething(t *testing.T) {
	m := NewMetricsReporter("ns", "sys")
	m.ReportRequest(context.Background(), time.Now(), "test", nil)
}
