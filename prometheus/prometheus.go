package prometheus

import (
	"net/http"
	"strings"

	"github.com/jukylin/esim/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Prometheus struct{}

func NewPrometheus(http_addr string, logger log.Logger) *Prometheus {

	prometheus := &Prometheus{}

	in := strings.Index(http_addr, ":")
	if in < 0 {
		http_addr = ":" + http_addr
	}
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Panicf(http.ListenAndServe(http_addr, nil).Error())
	}()
	logger.Infof("[prometheus] %s init success", http_addr)

	return prometheus
}

func NewNullProme() *Prometheus {

	prome := &Prometheus{}

	return prome
}
