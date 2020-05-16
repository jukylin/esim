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

// MonitorProxy wraps a RoundTripper.
type MonitorProxy struct {
	nextTransport http.RoundTripper

	logger log.Logger

	conf config.Config

	//use nethttp.Tracer
	tracer opentracing2.Tracer

	afterEvents []afterEvents

	name string
}

type afterEvents func(beginTime, endTime time.Time,
	res *http.Request, resp *http.Response)

type MonitorProxyOption func(c *MonitorProxy)

type MonitorProxyOptions struct{}

func NewMonitorProxy(options ...MonitorProxyOption) *MonitorProxy {

	MonitorProxy := &MonitorProxy{}

	for _, option := range options {
		option(MonitorProxy)
	}

	if MonitorProxy.conf == nil {
		MonitorProxy.conf = config.NewNullConfig()
	}

	if MonitorProxy.logger == nil {
		MonitorProxy.logger = log.NewLogger()
	}

	if MonitorProxy.tracer == nil {
		MonitorProxy.tracer = opentracing.NewTracer("http",
			MonitorProxy.logger)
	}

	MonitorProxy.registerAfterEvent()

	MonitorProxy.name = "monitor_proxy"

	return MonitorProxy
}

func (MonitorProxyOptions) WithConf(conf config.Config) MonitorProxyOption {
	return func(pt *MonitorProxy) {
		pt.conf = conf
	}
}

func (MonitorProxyOptions) WithLogger(logger log.Logger) MonitorProxyOption {
	return func(pt *MonitorProxy) {
		pt.logger = logger
	}
}

//use nethttp.Tracer
func (MonitorProxyOptions) WithTracer(tracer opentracing2.Tracer) MonitorProxyOption {
	return func(c *MonitorProxy) {
		c.tracer = tracer
	}
}

func (mp *MonitorProxy) NextProxy(tripper interface{}) {
	mp.nextTransport = tripper.(http.RoundTripper)
}

func (mp *MonitorProxy) ProxyName() string {
	return mp.name
}

// RoundTrip implements the RoundTripper interface.
func (mp *MonitorProxy) RoundTrip(req *http.Request) (*http.Response, error) {
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

func (mp *MonitorProxy) registerAfterEvent() {

	if mp.conf.GetBool("http_client_check_slow") {
		mp.afterEvents = append(mp.afterEvents, mp.slowHTTPRequest)
	}

	if mp.conf.GetBool("http_client_metrics") {
		mp.afterEvents = append(mp.afterEvents, mp.httpClientMetrice)
	}

	if mp.conf.GetBool("debug") {
		mp.afterEvents = append(mp.afterEvents, mp.debugHTTP)
	}
}

func (mp *MonitorProxy) after(beginTime time.Time, res *http.Request, resp *http.Response) {
	endTime := time.Now()
	for _, event := range mp.afterEvents {
		event(beginTime, endTime, res, resp)
	}
}

func (mp *MonitorProxy) slowHTTPRequest(beginTime, endTime time.Time,
	res *http.Request, resp *http.Response) {
	httpClientSlowTime := mp.conf.GetInt64("http_client_slow_time")

	if httpClientSlowTime != 0 {
		if endTime.Sub(beginTime) > time.Duration(httpClientSlowTime)*time.Millisecond {
			mp.logger.Warnf("slow http request [%s] ：%s", endTime.Sub(beginTime).String(),
				res.RequestURI)
		}
	}
}

func (mp *MonitorProxy) httpClientMetrice(beginTime, endTime time.Time,
	res *http.Request, resp *http.Response) {
	lab := prometheus.Labels{"url": res.URL.String(), "method": res.Method}
	httpTotal.With(lab).Inc()
	httpDuration.With(lab).Observe(endTime.Sub(beginTime).Seconds())
}

func (mp *MonitorProxy) debugHTTP(beginTime, endTime time.Time,
	req *http.Request, resp *http.Response) {
	mp.logger.Debugf("http [%d] [%s] %s ： %s", resp.StatusCode, endTime.Sub(beginTime).String(),
		req.Method, req.URL.String())
}
