package http

import (
	"net/http"
	"github.com/jukylin/esim/log"
)

type spyProxy struct {

	RoundTripWasCalled bool

	name string

	log log.Logger

	nextTransport http.RoundTripper
}

type spyProxyOption func(c *spyProxy)

type spyProxyOptions struct{}

func NewSpyProxy(logger log.Logger, name string) *spyProxy {
	spyProxy := &spyProxy{}

	if logger == nil{
		spyProxy.log = log.NewLogger()
	}else{
		spyProxy.log = logger
	}

	spyProxy.name = name

	return spyProxy
}

func (this *spyProxy) NextProxy(tripper interface{})  {
	this.nextTransport = tripper.(http.RoundTripper)
}

func (this *spyProxy) ProxyName() string {
	return this.name
}

// RoundTrip implements the RoundTripper interface.
func (this *spyProxy) RoundTrip(req *http.Request) (*http.Response, error) {
	if this.nextTransport == nil {
		this.nextTransport = http.DefaultTransport
	}

	this.RoundTripWasCalled = true
	this.log.Infof("%s was called", this.name)
	resp, err := this.nextTransport.RoundTrip(req)

	return resp, err
}