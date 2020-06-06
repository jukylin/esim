package middleware

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(requestTotal)
	prometheus.MustRegister(requestDuration)
}

// web_reqeust_total.
var requestTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "web_reqeust_total",
		Help: "Number of hello requests in total",
	},
	[]string{"method", "endpoint"},
)

// web_request_duration_seconds.
var requestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "web_request_duration_seconds",
		Help:    "web request duration distribution",
		Buckets: []float64{0.1, 0.3, 0.5, 0.7, 0.9, 1, 3, 5, 10, 30, 100},
	},
	[]string{"method", "endpoint"},
)
