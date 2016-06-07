package service

import (
	"encoding/json"
	_ "expvar"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type testExpvar struct {
	ErrorCounter      int     `json:"error_counter"`
	RequestCounter    int     `json:"request_counter"`
	RequestDuration99 float64 `json:"request_duration_ns_p99"`
}

func TestMetrics(t *testing.T) {
	prometheusServer := httptest.NewServer(prometheus.Handler())
	defer prometheusServer.Close()
	// Shitty default expvar package exports to the default MUX: https://github.com/golang/go/issues/15030
	expvarServer := httptest.NewServer(http.DefaultServeMux)
	defer expvarServer.Close()

	m := NewMetricsReporter("ns", "sys")
	m.ReportRequest(context.Background(), time.Now().Add(-time.Second), "test1", nil)
	m.ReportRequest(context.Background(), time.Now().Add(-time.Millisecond), "test1", nil)
	m.ReportRequest(context.Background(), time.Now().Add(-time.Second), "test2", nil)
	m.ReportRequest(context.Background(), time.Now().Add(-time.Millisecond), "test2", xerror.New("some-error"))

	// Verify Prometheus metrics.
	resp, err := http.Get(prometheusServer.URL)
	assert.Nil(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, resp.Body.Close())
	prometheusMetrics := parsePrometheus(string(body))
	assertMetric(t, prometheusMetrics, "ns_sys_request_duration_ns_count{action=\"test1\"}", "2")
	assertMetric(t, prometheusMetrics, "ns_sys_request_duration_ns_count{action=\"test2\"}", "2")
	assertMetric(t, prometheusMetrics, "ns_sys_request_counter{action=\"test1\"}", "2")
	assertMetric(t, prometheusMetrics, "ns_sys_request_counter{action=\"test2\"}", "2")
	assertMetric(t, prometheusMetrics, "ns_sys_error_counter{action=\"test2\"}", "1")

	// Verify expvar metrics.
	resp, err = http.Get(expvarServer.URL + "/debug/vars")
	assert.Nil(t, err)
	body, err = ioutil.ReadAll(resp.Body)
	assert.Nil(t, resp.Body.Close())
	expvarMetrics := &testExpvar{}
	assert.Nil(t, json.Unmarshal(body, expvarMetrics))
	assert.Equal(t, 1, expvarMetrics.ErrorCounter)
	assert.Equal(t, 4, expvarMetrics.RequestCounter)
	assert.True(t, expvarMetrics.RequestDuration99 > float64(time.Second))
}

func parsePrometheus(body string) map[string]string {
	metrics := make(map[string]string)
	for _, line := range strings.Split(body, "\n") {
		if len(strings.TrimSpace(line)) > 0 && !strings.HasPrefix(line, "#") {
			kv := strings.SplitAfterN(line, " ", 2)
			metrics[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return metrics
}

func assertMetric(t *testing.T, metrics map[string]string, key, expectedValue string) {
	value, ok := metrics[key]
	assert.True(t, ok)
	assert.Equal(t, expectedValue, value)
}
