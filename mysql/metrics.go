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


var mysqlStats = prometheus.NewGaugeVec(
	prometheus.GaugeOpts {
		Namespace: "mysql_stats",
		Subsystem: "blob_storage",
		Name:      "ops_queued",
		Help:      "Number of blob storage operations waiting to be processed.",
	},
	[]string{"db", "stats"},
)

func init() {
	prometheus.MustRegister(mysqlTotal)
	prometheus.MustRegister(mysqlDuration)
	prometheus.MustRegister(mysqlStats)
}
