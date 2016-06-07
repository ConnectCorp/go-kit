package service

import (
	"github.com/ConnectCorp/go-kit/common/middleware"
	"github.com/go-kit/kit/endpoint"
	kitmetrics "github.com/go-kit/kit/metrics"
	kitexpvar "github.com/go-kit/kit/metrics/expvar"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"time"
)

const (
	ctxLabelMetricsReporter = "metricsReporter"
	requestDurationLabel    = "request_duration_ns"
	requestCounterLabel     = "request_counter"
	errorCounterLabel       = "error_counter"
)

// MetricsReporter is an interface that allows to report standard metrics for a request.
type MetricsReporter interface {
	ReportRequest(ctx context.Context, startTime time.Time, action string, err error)
}

type metricsReporter struct {
	namespace             string
	system                string
	requestDurationMetric kitmetrics.TimeHistogram
	requestCounterMetric  kitmetrics.Counter
	errorCounterMetric    kitmetrics.Counter
}

// ReportRequest implements the MetricsReporter interface.
func (m *metricsReporter) ReportRequest(_ context.Context, startTime time.Time, action string, err error) {
	m.requestDurationMetric.With(kitmetrics.Field{Key: "action", Value: action}).Observe(time.Since(startTime))
	m.requestCounterMetric.Add(1)
	if err != nil {
		m.errorCounterMetric.Add(1)
	}
}

// NewMetricsReporter creates a new MetricsReporter that targets expvar and Prometheus.
func NewMetricsReporter(namespace, system string) MetricsReporter {
	return &metricsReporter{
		namespace:             namespace,
		system:                system,
		requestDurationMetric: makeRequestDurationMetric(namespace, system, requestDurationLabel),
		requestCounterMetric:  makeRequestCounterMetric(namespace, system, requestCounterLabel),
		errorCounterMetric:    makeErrorCounterMetric(namespace, system, errorCounterLabel),
	}
}

func makeRequestDurationMetric(namespace, system, label string) kitmetrics.TimeHistogram {
	return kitmetrics.NewTimeHistogram(time.Nanosecond, kitmetrics.NewMultiHistogram(
		label,
		kitexpvar.NewHistogram(label, 0, 5e9, 1, 50, 95, 99),
		kitprometheus.NewSummary(prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: system,
			Name:      label,
			Help:      "Request duration in nanoseconds.",
		}, []string{"action"})))
}

func makeRequestCounterMetric(namespace, system, label string) kitmetrics.Counter {
	return kitmetrics.NewMultiCounter(
		label,
		kitexpvar.NewCounter(label),
		kitprometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: system,
			Name:      label,
			Help:      "Total number of requests.",
		}, []string{"requests"}))
}

func makeErrorCounterMetric(namespace, system, label string) kitmetrics.Counter {
	return kitmetrics.NewMultiCounter(
		label,
		kitexpvar.NewCounter(label),
		kitprometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: system,
			Name:      label,
			Help:      "Total number of errors.",
		}, []string{"errors"}))
}

// NewMetricsMiddleware creates a new standard metrics middleware for a Go microservice.
func NewMetricsMiddleware(system, namespace string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (resp interface{}, err error) {
			metricsReporter := ctxMetricsReporter(ctx)
			defer func(startTime time.Time) {
				metricsReporter.ReportRequest(ctx, startTime, ctxRequestPath(ctx), err)
			}(time.Now())
			return next(ctx, request)
		}
	}
}

func ctxWithMetricsReporter(ctx context.Context, metricsReporter middleware.MetricsReporter) context.Context {
	return context.WithValue(ctx, ctxLabelMetricsReporter, metricsReporter)
}

func ctxMetricsReporter(ctx context.Context) middleware.MetricsReporter {
	v := ctx.Value(ctxLabelMetricsReporter)
	if v == nil {
		return nil
	}
	return v.(middleware.MetricsReporter)
}
