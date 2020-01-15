package prometheus

import (
	"net/http"
	"strings"

	//"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/config"
)

type Prometheus struct {}

func NewPrometheus(conf config.Config, log log.Logger) *Prometheus {

	prometheus := &Prometheus{}

	prometheus_http_addr := conf.GetString("prometheus_http_addr")

	in := strings.Index(prometheus_http_addr, ":")
	if in < 0 {
		prometheus_http_addr = ":"+prometheus_http_addr
	}
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Panicf(http.ListenAndServe(prometheus_http_addr, nil).Error())
	}()
	log.Infof("[prometheus] %s init success", prometheus_http_addr)

	return prometheus
}


func NewNullProme() *Prometheus {

	prome := &Prometheus{}

	return prome
}
