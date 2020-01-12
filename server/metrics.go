
package server

import (
	"github.com/prometheus/client_golang/prometheus"
	)

// 初始化 web_reqeust_total
var requestTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "web_reqeust_total",
		Help: "Number of hello requests in total",
	},
	// 设置两个标签 请求方法和 路径 对请求总次数在两个
	[]string{"method", "endpoint"},
)

// web_request_duration_seconds
var requestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "web_request_duration_seconds",
		Help:    "web request duration distribution",
		Buckets: []float64{0.1, 0.3, 0.5, 0.7, 0.9, 1, 3, 5, 10, 30, 100},
	},
	[]string{"method", "endpoint"},
)

func init() {
	prometheus.MustRegister(requestTotal)
	prometheus.MustRegister(requestDuration)
}
