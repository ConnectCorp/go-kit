package service

import (
	"github.com/go-kit/kit/endpoint"
	kitmetrics "github.com/go-kit/kit/metrics"
	kitdogstatsd "github.com/go-kit/kit/metrics/dogstatsd"
	kitexpvar "github.com/go-kit/kit/metrics/expvar"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"time"
)

const (
	commonMetricsNamespace = "connect"
	requestDurationLabel   = "request_duration_ms"
	requestCounterLabel    = "request_counter"
	errorCounterLabel      = "error_counter"
)

// MetricsReporter is an interface that allows to report standard metrics for a request.
type MetricsReporter interface {
	ReportRequest(ctx context.Context, startTime time.Time, action string, err error)
}

type metricsReporter struct {
	namespace             string
	system                string
	dogstatsdEmitter      *kitdogstatsd.Emitter
	requestDurationMetric kitmetrics.TimeHistogram
	requestCounterMetric  kitmetrics.Counter
	errorCounterMetric    kitmetrics.Counter
}

// ReportRequest implements the MetricsReporter interface.
func (m *metricsReporter) ReportRequest(_ context.Context, startTime time.Time, action string, err error) {
	m.requestDurationMetric.With(kitmetrics.Field{Key: "action", Value: action}).Observe(time.Since(startTime))
	m.requestCounterMetric.With(kitmetrics.Field{Key: "action", Value: action}).Add(1)
	if err != nil {
		m.errorCounterMetric.With(kitmetrics.Field{Key: "action", Value: action}).Add(1)
	}
}

// NewMetricsReporter creates a new MetricsReporter that targets expvar and Prometheus.
func NewMetricsReporter(namespace, system string, dogstatsdEmitter *kitdogstatsd.Emitter) MetricsReporter {
	return &metricsReporter{
		namespace:             namespace,
		system:                system,
		dogstatsdEmitter:      dogstatsdEmitter,
		requestDurationMetric: makeRequestDurationMetric(namespace, system, requestDurationLabel, dogstatsdEmitter),
		requestCounterMetric:  makeRequestCounterMetric(namespace, system, requestCounterLabel, dogstatsdEmitter),
		errorCounterMetric:    makeErrorCounterMetric(namespace, system, errorCounterLabel, dogstatsdEmitter),
	}
}

func makeRequestDurationMetric(namespace, system, label string, dogstatsdEmitter *kitdogstatsd.Emitter) kitmetrics.TimeHistogram {
	histograms := []kitmetrics.Histogram{
		kitexpvar.NewHistogram(label, 0, 5e9, 1, 50, 95, 99),
		kitprometheus.NewSummary(prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: system,
			Name:      label,
			Help:      "Request duration in nanoseconds.",
		}, []string{"action"}),
	}

	if dogstatsdEmitter != nil {
		histograms = append(histograms, dogstatsdEmitter.NewHistogram(label))
	}

	return kitmetrics.NewTimeHistogram(time.Millisecond, kitmetrics.NewMultiHistogram(label, histograms...))
}

func makeRequestCounterMetric(namespace, system, label string, dogstatsdEmitter *kitdogstatsd.Emitter) kitmetrics.Counter {
	counters := []kitmetrics.Counter{
		kitexpvar.NewCounter(label),
		kitprometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: system,
			Name:      label,
			Help:      "Total number of requests.",
		}, []string{"action"}),
	}

	if dogstatsdEmitter != nil {
		counters = append(counters, dogstatsdEmitter.NewCounter(label))
	}

	return kitmetrics.NewMultiCounter(label, counters...)
}

func makeErrorCounterMetric(namespace, system, label string, dogstatsdEmitter *kitdogstatsd.Emitter) kitmetrics.Counter {
	counters := []kitmetrics.Counter{
		kitexpvar.NewCounter(label),
		kitprometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: system,
			Name:      label,
			Help:      "Total number of errors.",
		}, []string{"action"}),
	}

	if dogstatsdEmitter != nil {
		counters = append(counters, dogstatsdEmitter.NewCounter(label))
	}

	return kitmetrics.NewMultiCounter(label, counters...)
}

// NewMetricsMiddleware creates a new standard metrics middleware for a Go microservice.
func NewMetricsMiddleware(metricsReporter MetricsReporter) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (resp interface{}, err error) {
			defer func(startTime time.Time) {
				metricsReporter.ReportRequest(ctx, startTime, ctxRequestPath(ctx), err)
			}(time.Now())
			return next(ctx, request)
		}
	}
}
