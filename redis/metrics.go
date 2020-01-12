package redis

import (
	"github.com/prometheus/client_golang/prometheus"
	)

var redisTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "redis_total",
		Help: "Number of hello requests in total",
	},
	[]string{"cmd"},
)

// redis_duration_seconds
var redisDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "redis_duration_seconds",
		Help:    "redis duration distribution",
		Buckets: []float64{0.001, 0.003, 0.005, 0.007, 0.01},
	},
	[]string{"cmd"},
)

func init() {
	prometheus.MustRegister(redisTotal)
	prometheus.MustRegister(redisDuration)
}
