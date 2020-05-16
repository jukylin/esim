package http

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

var logger log.Logger

var (
	host1 = "127.0.0.1"
	host2 = "127.0.0.2"
)

func TestMain(m *testing.M) {
	loggerOptions := log.LoggerOptions{}
	logger = log.NewLogger(loggerOptions.WithDebug(true))

	code := m.Run()

	os.Exit(code)
}

func TestMulLevelRoundTrip(t *testing.T) {

	clientOptions := ClientOptions{}
	httpClient := NewClient(
		clientOptions.WithLogger(logger),
		clientOptions.WithProxy(
			func() interface{} {
				return newSpyProxy(logger, "spyProxy1")
			},
			func() interface{} {
				return newSpyProxy(logger, "spyProxy2")
			},
			func() interface{} {
				stubsProxyOptions := StubsProxyOptions{}
				stubsProxy := newStubsProxy(
					stubsProxyOptions.WithRespFunc(func(request *http.Request) *http.Response {
						resp := &http.Response{}
						if request.URL.String() == host1 {
							resp.StatusCode = 200
						} else if request.URL.String() == host2 {
							resp.StatusCode = 300
						}

						resp.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))

						return resp
					}),
					stubsProxyOptions.WithName("stubsProxy1"),
					stubsProxyOptions.WithLogger(logger),
				)

				return stubsProxy
			},
		),
	)

	testCases := []struct {
		behavior string
		url      string
		result   int
	}{
		{"127.0.0.1:200", host1, 200},
		{"127.0.0.2:300", host2, 300},
	}

	ctx := context.Background()

	for _, test := range testCases {
		t.Run(test.behavior, func(t *testing.T) {
			resp, err := httpClient.Get(ctx, test.url)
			resp.Body.Close()

			assert.Nil(t, err)
			assert.Equal(t, test.result, resp.StatusCode)
		})
	}
}

func TestMonitorProxy(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("http_client_metrics", true)

	monitorProxyOptions := MonitorProxyOptions{}

	clientOptions := ClientOptions{}
	httpClient := NewClient(
		clientOptions.WithLogger(logger),
		clientOptions.WithProxy(
			func() interface{} {
				return NewMonitorProxy(
					monitorProxyOptions.WithConf(memConfig),
					monitorProxyOptions.WithLogger(logger))
			},
			func() interface{} {
				stubsProxyOptions := StubsProxyOptions{}
				stubsProxy := newStubsProxy(
					stubsProxyOptions.WithRespFunc(func(request *http.Request) *http.Response {
						resp := &http.Response{}
						if request.URL.String() == host1 {
							resp.StatusCode = 200
						} else if request.URL.String() == host2 {
							resp.StatusCode = 300
						}

						resp.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))

						return resp
					}),
					stubsProxyOptions.WithName("stubsProxy1"),
					stubsProxyOptions.WithLogger(logger),
				)

				return stubsProxy
			},
		),
	)

	ctx := context.Background()
	resp, err := httpClient.Get(ctx, host1)
	resp.Body.Close()

	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	resp, err = httpClient.Get(ctx, host2)
	resp.Body.Close()

	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, 300)

	lab := prometheus.Labels{"url": host1, "method": "GET"}
	c, _ := httpTotal.GetMetricWith(lab)
	metric := &io_prometheus_client.Metric{}
	err = c.Write(metric)
	assert.Nil(t, err)

	assert.Equal(t, float64(1), metric.Counter.GetValue())
}

func TestTimeoutProxy(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("http_client_check_slow", true)
	memConfig.Set("http_client_slow_time", 5)

	monitorProxyOptions := MonitorProxyOptions{}

	clientOptions := ClientOptions{}
	httpClient := NewClient(
		clientOptions.WithLogger(logger),
		clientOptions.WithProxy(
			func() interface{} {
				return NewMonitorProxy(
					monitorProxyOptions.WithConf(memConfig),
					monitorProxyOptions.WithLogger(logger))
			},
			func() interface{} {
				return newSlowProxy(logger, "slowProxy")
			},
			func() interface{} {
				stubsProxyOptions := StubsProxyOptions{}
				stubsProxy := newStubsProxy(
					stubsProxyOptions.WithRespFunc(func(request *http.Request) *http.Response {
						resp := &http.Response{}
						if request.URL.String() == host1 {
							resp.StatusCode = 200
						} else if request.URL.String() == host2 {
							resp.StatusCode = 300
						}

						resp.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))

						return resp
					}),
					stubsProxyOptions.WithName("stubsProxy1"),
					stubsProxyOptions.WithLogger(logger),
				)

				return stubsProxy
			},
		),
	)

	ctx := context.Background()
	resp, err := httpClient.Get(ctx, host1)
	resp.Body.Close()

	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, 200)
}
