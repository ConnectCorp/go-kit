package middleware

import (
	kitmetrics "github.com/go-kit/kit/metrics"
	kitexpvar "github.com/go-kit/kit/metrics/expvar"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"time"
)

type MetricsReporter interface {
	ReportRequest(ctx context.Context, startTime time.Time, action string, err error)
}

type metrics struct {
	namespace             string
	system                string
	requestDurationMetric kitmetrics.TimeHistogram
	requestCounterMetric  kitmetrics.Counter
	errorCounterMetric    kitmetrics.Counter
}

const (
	requestDurationLabel = "request_duration_ns"
	requestCounterLabel  = "request_counter"
	errorCounterLabel    = "error_counter"
)

func NewPrometheusMetrics(namespace, system string) *metrics {
	return &metrics{
		system:                system,
		namespace:             namespace,
		requestDurationMetric: MakeRequestDurationMetric(namespace, system, requestDurationLabel),
		requestCounterMetric:  MakeRequestCounterMetric(namespace, system, requestCounterLabel),
		errorCounterMetric:    MakeErrorCounterMetric(namespace, system, errorCounterLabel),
	}
}

func MakeRequestDurationMetric(namespace, system, label string) kitmetrics.TimeHistogram {
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

func MakeRequestCounterMetric(namespace, system, label string) kitmetrics.Counter {
	return kitmetrics.NewMultiCounter(
		label,
		kitexpvar.NewCounter(label),
		kitprometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: system,
			Name:      label,
			Help:      "Number of requests processed.",
		}, []string{"requests"}))
}

func MakeErrorCounterMetric(namespace, system, label string) kitmetrics.Counter {
	return kitmetrics.NewMultiCounter(
		label,
		kitexpvar.NewCounter(label),
		kitprometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: system,
			Name:      label,
			Help:      "Number of errors processed.",
		}, []string{"errors"}))
}

func (m *metrics) ReportRequest(ctx context.Context, startTime time.Time, action string, err error) {

	m.requestDurationMetric.
		With(kitmetrics.Field{Key: "action", Value: action}).
		Observe(time.Since(startTime))

	m.requestCounterMetric.Add(1)

	if err != nil {
		m.errorCounterMetric.Add(1)
	}
}
