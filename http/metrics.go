package http

import (
	"github.com/prometheus/client_golang/prometheus"
	)

var httpTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_total",
		Help: "Number of total",
	},
	[]string{"url", "method"},
)

var httpDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_duration_seconds",
		Help:    "http duration distribution",
		Buckets: []float64{0.02, 0.08, 0.15, 0.5, 1, 3},
	},
	[]string{"url", "method"},
)

func init() {
	prometheus.MustRegister(httpTotal)
	prometheus.MustRegister(httpDuration)
}
