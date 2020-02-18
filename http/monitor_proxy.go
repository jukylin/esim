package http

import (
	"net/http"
	"time"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/opentracing"
)

// monitorProxy wraps a RoundTripper.
type monitorProxy struct {
	nextTransport http.RoundTripper

	log log.Logger

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

	if monitorProxy.log == nil {
		monitorProxy.log = log.NewLogger()
	}

	if monitorProxy.tracer == nil {
		monitorProxy.tracer = opentracing.NewTracer("http",
			monitorProxy.log)
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

func (MonitorProxyOptions) WithLogger(log log.Logger) MonitorProxyOption {
	return func(pt *monitorProxy) {
		pt.log = log
	}
}

//use nethttp.Tracer
func (MonitorProxyOptions) WithTracer(tracer opentracing2.Tracer) MonitorProxyOption {
	return func(c *monitorProxy) {
		c.tracer = tracer
	}
}

func (this *monitorProxy) NextProxy(tripper interface{})  {
	this.nextTransport = tripper.(http.RoundTripper)
}

func (this *monitorProxy) ProxyName() string {
	return this.name
}

// RoundTrip implements the RoundTripper interface.
func (this *monitorProxy) RoundTrip(req *http.Request) (*http.Response, error) {
	if this.nextTransport == nil {
		this.nextTransport = http.DefaultTransport
	}

	beginTime := time.Now()

	var resp *http.Response
	var err error

	if this.conf.GetBool("http_client_tracer") == true {
		req, ht := nethttp.TraceRequest(this.tracer, req)

		transport := nethttp.Transport{}
		transport.RoundTripper = this.nextTransport
		resp, err = transport.RoundTrip(req)

		ht.Finish()
	}else{
		resp, err = this.nextTransport.RoundTrip(req)
	}

	if err != nil {
		return resp, err
	}

	this.after(beginTime, req, resp)


	return resp, nil
}

func (this *monitorProxy) registerAfterEvent() {

	if this.conf.GetBool("http_client_check_slow") == true {
		this.afterEvents = append(this.afterEvents, this.slowHttpRequest)
	}

	if this.conf.GetBool("http_client_metrics") == true {
		this.afterEvents = append(this.afterEvents, this.httpClientMetrice)
	}

	if this.conf.GetBool("debug") == true {
		this.afterEvents = append(this.afterEvents, this.debugHttp)
	}
}

func (this *monitorProxy) after(beginTime time.Time, res *http.Request, resp *http.Response) {
	endTime := time.Now()
	for _, event := range this.afterEvents {
		event(beginTime, endTime, res, resp)
	}
}


func (this *monitorProxy) slowHttpRequest(beginTime time.Time, endTime time.Time,
	res *http.Request, resp *http.Response) {
	http_client_slow_time := this.conf.GetInt64("http_client_slow_time")

	if http_client_slow_time != 0 {
		if endTime.Sub(beginTime) > time.Duration(http_client_slow_time)*time.Millisecond {
			this.log.Warnf("slow http request [%s] ：%s", endTime.Sub(beginTime).String(),
				res.RequestURI)
		}
	}
}

func (this *monitorProxy) httpClientMetrice(beginTime time.Time, endTime time.Time,
	res *http.Request, resp *http.Response) {
	lab := prometheus.Labels{"url": res.URL.String(), "method": res.Method}
	httpTotal.With(lab).Inc()
	httpDuration.With(lab).Observe(endTime.Sub(beginTime).Seconds())
}

func (this *monitorProxy) debugHttp(beginTime time.Time, endTime time.Time,
	req *http.Request, resp *http.Response) {
	this.log.Debugf("http [%d] [%s] %s ： %s", resp.StatusCode, endTime.Sub(beginTime).String(),
		req.Method, req.URL.String())
}
