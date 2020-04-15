package prometheus

import (
	"net/http"
	"strings"

	"github.com/jukylin/esim/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Prometheus struct{}

func NewPrometheus(httpAddr string, logger log.Logger) *Prometheus {

	prometheus := &Prometheus{}

	in := strings.Index(httpAddr, ":")
	if in < 0 {
		httpAddr = ":" + httpAddr
	}
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Panicf(http.ListenAndServe(httpAddr, nil).Error())
	}()
	logger.Infof("[prometheus] %s init success", httpAddr)

	return prometheus
}

func NewNullProme() *Prometheus {

	prome := &Prometheus{}

	return prome
}
