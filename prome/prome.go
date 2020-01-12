package prome

import (
	"net/http"
	"strings"

	//"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/config"
)

type Prome struct {}

func NewProme(conf config.Config, log log.Logger) *Prome {

	prome := &Prome{}

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

	return prome
}


func NewNullProme() *Prome {

	prome := &Prome{}

	return prome
}
