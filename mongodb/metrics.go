package mongodb

import (
	"github.com/prometheus/client_golang/prometheus"
)

var mongodbTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mongodb_total",
		Help: "Number of total",
	},
	[]string{"command"},
)

var mongodbDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "mongodb_duration_seconds",
		Help:    "mongodb duration distribution",
		Buckets: []float64{0.02, 0.08, 0.15, 0.5, 1, 3},
	},
	[]string{"command"},
)

var mongodbErrTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mongodb_err_total",
		Help: "Number of err total",
	},
	[]string{"command"},
)

var mongodbPoolTypes = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mongodb_pool_total",
		Help: "Number of pool's types total",
	},
	[]string{"type"},
)

func init() {
	prometheus.MustRegister(mongodbTotal)
	prometheus.MustRegister(mongodbDuration)
	prometheus.MustRegister(mongodbErrTotal)
	prometheus.MustRegister(mongodbPoolTypes)
}
