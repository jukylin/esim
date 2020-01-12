package mysql

import (
	"github.com/prometheus/client_golang/prometheus"
	)

var mysqlTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mysql_total",
		Help: "Number of total",
	},
	[]string{"sql"},
)

var mysqlDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "mysql_duration_seconds",
		Help:    "mysql duration distribution",
		Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1},
	},
	[]string{"sql"},
)

func init() {
	prometheus.MustRegister(mysqlTotal)
	prometheus.MustRegister(mysqlDuration)
}
