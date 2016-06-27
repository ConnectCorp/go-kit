package service

import (
	"bytes"
	"encoding/json"
	_ "expvar"
	kitdogstatsd "github.com/go-kit/kit/metrics/dogstatsd"
	"github.com/go-kit/kit/util/conn"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"gopkg.in/ibrt/go-xerror.v2/xerror"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
)

type testExpvar struct {
	ErrorCounter      int     `json:"error_counter"`
	RequestCounter    int     `json:"request_counter"`
	RequestDuration99 float64 `json:"request_duration_ms_p99"`
}

func TestMetrics(t *testing.T) {
	prometheusServer := httptest.NewServer(prometheus.Handler())
	defer prometheusServer.Close()
	// Shitty default expvar package exports to the default MUX: https://github.com/golang/go/issues/15030
	expvarServer := httptest.NewServer(http.DefaultServeMux)
	defer expvarServer.Close()

	dogstatsdBuffer := &syncbuf{buf: &bytes.Buffer{}}
	dogstatsdEmitter := kitdogstatsd.NewEmitterDial(mockDialer(dogstatsdBuffer), "", "", "test_", time.Millisecond, log.NewNopLogger())

	m := NewMetricsReporter("ns", "sys", dogstatsdEmitter)
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
	assertMetric(t, prometheusMetrics, "ns_sys_request_duration_ms_count{action=\"test1\"}", "2")
	assertMetric(t, prometheusMetrics, "ns_sys_request_duration_ms_count{action=\"test2\"}", "2")
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
	assert.True(t, expvarMetrics.RequestDuration99 > float64(time.Second/time.Millisecond))

	// Ensure that dogstatsd metrics are emitted.
	time.Sleep(2 * time.Millisecond)

	// Verify dogstatsd metrics.
	dogstatsdMetrics := strings.Split(dogstatsdBuffer.String(), "\n")
	assert.Contains(t, dogstatsdMetrics, "test_request_duration_ms:1000|ms|#action:test1")
	assert.Contains(t, dogstatsdMetrics, "test_request_duration_ms:1|ms|#action:test1")
	assert.Contains(t, dogstatsdMetrics, "test_request_counter:1|c|#action:test1")
	assert.Contains(t, dogstatsdMetrics, "test_request_counter:1|c|#action:test1")
	assert.Contains(t, dogstatsdMetrics, "test_request_duration_ms:1000|ms|#action:test2")
	assert.Contains(t, dogstatsdMetrics, "test_request_counter:1|c|#action:test2")
	assert.Contains(t, dogstatsdMetrics, "test_request_duration_ms:1|ms|#action:test2")
	assert.Contains(t, dogstatsdMetrics, "test_request_counter:1|c|#action:test2")
	assert.Contains(t, dogstatsdMetrics, "test_error_counter:1|c|#action:test2")
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

func mockDialer(buf *syncbuf) conn.Dialer {
	return func(net, addr string) (net.Conn, error) {
		return &mockConn{buf}, nil
	}
}

type syncbuf struct {
	mtx sync.Mutex
	buf *bytes.Buffer
}

func (s *syncbuf) Write(p []byte) (int, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.buf.Write(p)
}

func (s *syncbuf) String() string {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.buf.String()
}

func (s *syncbuf) Reset() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.buf.Reset()
}

type mockConn struct {
	buf *syncbuf
}

func (c *mockConn) Read(b []byte) (n int, err error) {
	panic("not implemented")
}

func (c *mockConn) Write(b []byte) (n int, err error) {
	return c.buf.Write(b)
}

func (c *mockConn) Close() error {
	panic("not implemented")
}

func (c *mockConn) LocalAddr() net.Addr {
	panic("not implemented")
}

func (c *mockConn) RemoteAddr() net.Addr {
	panic("not implemented")
}

func (c *mockConn) SetDeadline(t time.Time) error {
	panic("not implemented")
}

func (c *mockConn) SetReadDeadline(t time.Time) error {
	panic("not implemented")
}

func (c *mockConn) SetWriteDeadline(t time.Time) error {
	panic("not implemented")
}
