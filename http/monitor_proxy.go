package http

import (
	"net/http"
	"time"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/opentracing"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
)

// monitorProxy wraps a RoundTripper.
type monitorProxy struct {
	nextTransport http.RoundTripper

	logger log.Logger

	conf config.Config

	//use nethttp.Tracer
	tracer opentracing2.Tracer

	afterEvents []afterEvents

	name string
}

type afterEvents func(beginTime time.Time, endTime time.Time,
	res *http.Request, resp *http.Response)

type MonitorProxyOption func(c *monitorProxy)

type MonitorProxyOptions struct{}

func NewMonitorProxy(options ...MonitorProxyOption) *monitorProxy {

	monitorProxy := &monitorProxy{}

	for _, option := range options {
		option(monitorProxy)
	}

	if monitorProxy.conf == nil {
		monitorProxy.conf = config.NewNullConfig()
	}

	if monitorProxy.logger == nil {
		monitorProxy.logger = log.NewLogger()
	}

	if monitorProxy.tracer == nil {
		monitorProxy.tracer = opentracing.NewTracer("http",
			monitorProxy.logger)
	}

	monitorProxy.registerAfterEvent()

	monitorProxy.name = "monitor_proxy"

	return monitorProxy
}

func (MonitorProxyOptions) WithConf(conf config.Config) MonitorProxyOption {
	return func(pt *monitorProxy) {
		pt.conf = conf
	}
}

func (MonitorProxyOptions) WithLogger(logger log.Logger) MonitorProxyOption {
	return func(pt *monitorProxy) {
		pt.logger = logger
	}
}

//use nethttp.Tracer
func (MonitorProxyOptions) WithTracer(tracer opentracing2.Tracer) MonitorProxyOption {
	return func(c *monitorProxy) {
		c.tracer = tracer
	}
}

func (mp *monitorProxy) NextProxy(tripper interface{}) {
	mp.nextTransport = tripper.(http.RoundTripper)
}

func (mp *monitorProxy) ProxyName() string {
	return mp.name
}

// RoundTrip implements the RoundTripper interface.
func (mp *monitorProxy) RoundTrip(req *http.Request) (*http.Response, error) {
	if mp.nextTransport == nil {
		mp.nextTransport = http.DefaultTransport
	}

	mp.logger.Debugc(req.Context(), "Url : %s", req.URL)

	beginTime := time.Now()

	var resp *http.Response
	var err error

	if mp.conf.GetBool("http_client_tracer") {
		req, ht := nethttp.TraceRequest(mp.tracer, req)

		transport := nethttp.Transport{}
		transport.RoundTripper = mp.nextTransport
		resp, err = transport.RoundTrip(req)

		ht.Finish()
	} else {
		resp, err = mp.nextTransport.RoundTrip(req)
	}

	if err != nil {
		return resp, err
	}

	mp.after(beginTime, req, resp)

	return resp, nil
}

func (mp *monitorProxy) registerAfterEvent() {

	if mp.conf.GetBool("http_client_check_slow") {
		mp.afterEvents = append(mp.afterEvents, mp.slowHttpRequest)
	}

	if mp.conf.GetBool("http_client_metrics") {
		mp.afterEvents = append(mp.afterEvents, mp.httpClientMetrice)
	}

	if mp.conf.GetBool("debug") {
		mp.afterEvents = append(mp.afterEvents, mp.debugHttp)
	}
}

func (mp *monitorProxy) after(beginTime time.Time, res *http.Request, resp *http.Response) {
	endTime := time.Now()
	for _, event := range mp.afterEvents {
		event(beginTime, endTime, res, resp)
	}
}

func (mp *monitorProxy) slowHttpRequest(beginTime time.Time, endTime time.Time,
	res *http.Request, resp *http.Response) {
	httpClientSlowTime := mp.conf.GetInt64("http_client_slow_time")

	if httpClientSlowTime != 0 {
		if endTime.Sub(beginTime) > time.Duration(httpClientSlowTime) * time.Millisecond {
			mp.logger.Warnf("slow http request [%s] ：%s", endTime.Sub(beginTime).String(),
				res.RequestURI)
		}
	}
}

func (mp *monitorProxy) httpClientMetrice(beginTime time.Time, endTime time.Time,
	res *http.Request, resp *http.Response) {
	lab := prometheus.Labels{"url": res.URL.String(), "method": res.Method}
	httpTotal.With(lab).Inc()
	httpDuration.With(lab).Observe(endTime.Sub(beginTime).Seconds())
}

func (mp *monitorProxy) debugHttp(beginTime time.Time, endTime time.Time,
	req *http.Request, resp *http.Response) {
	mp.logger.Debugf("http [%d] [%s] %s ： %s", resp.StatusCode, endTime.Sub(beginTime).String(),
		req.Method, req.URL.String())
}
