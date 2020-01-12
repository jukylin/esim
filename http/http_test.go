package http

import (
	"testing"
	"github.com/jukylin/esim/log"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_model/go"
)

func TestMulLevelRoundTrip(t *testing.T)  {

	logger := log.NewLogger()

	clientOptions := ClientOptions{}
	httpClient := NewHttpClient(
		clientOptions.WithLogger(log.NewLogger()),
		clientOptions.WithProxy(
			func() interface {} {
				return NewSpyProxy(logger, "spyProxy1")
			},
			func() interface {} {
				return NewSpyProxy(logger, "spyProxy2")
			},
			func() interface {} {
				return NewStubsProxy(logger, "stubsProxy1")
			},
			),
	)

	testCases := []struct{
		behavior string
		url string
		result int
	}{
		{"127.0.0.1:200", "127.0.0.1", 200},
		{"127.0.0.2:300", "127.0.0.2", 300},
	}

	ctx := context.Background()

	for _, test := range testCases{
		t.Run(test.behavior, func(t *testing.T) {
			resp, err := httpClient.Get(ctx, test.url)
			resp.Body.Close()

			assert.Nil(t, err)
			assert.Equal(t, test.result, resp.StatusCode)
		})
	}
}


func TestMonitorProxy(t *testing.T)  {
	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("http_client_metrics", true)

	monitorProxyOptions := MonitorProxyOptions{}

	clientOptions := ClientOptions{}
	httpClient := NewHttpClient(
		clientOptions.WithLogger(logger),
		clientOptions.WithProxy(
			func() interface {} {
				return NewMonitorProxy(
					monitorProxyOptions.WithConf(memConfig),
					monitorProxyOptions.WithLogger(logger))
			},
			func() interface {} {
				return NewStubsProxy(logger, "stubsProxy1")
			},
		),
	)

	ctx := context.Background()
	resp, err := httpClient.Get(ctx, "127.0.0.1")
	resp.Body.Close()

	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	resp, err = httpClient.Get(ctx, "127.0.0.2")
	resp.Body.Close()

	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, 300)

	lab := prometheus.Labels{"url": "127.0.0.1", "method": "GET"}
	c, _ := httpTotal.GetMetricWith(lab)
	metric := &io_prometheus_client.Metric{}
	c.Write(metric)
	assert.Equal(t, float64(1), metric.Counter.GetValue())
}



func TestTimeoutProxy(t *testing.T)  {
	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("http_client_check_slow", true)
	memConfig.Set("http_client_slow_time", 5)

	monitorProxyOptions := MonitorProxyOptions{}

	clientOptions := ClientOptions{}
	httpClient := NewHttpClient(
		clientOptions.WithLogger(logger),
		clientOptions.WithProxy(
			func() interface {} {
				return NewMonitorProxy(
					monitorProxyOptions.WithConf(memConfig),
					monitorProxyOptions.WithLogger(logger))
			},
			func() interface {} {
				return NewSlowProxy(logger, "slowProxy")
			},
			func() interface {} {
				return NewStubsProxy(logger, "stubsProxy1")
			},
		),
	)

	ctx := context.Background()
	resp, err := httpClient.Get(ctx, "127.0.0.1")
	resp.Body.Close()

	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, 200)
}