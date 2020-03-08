package http

import (
	"github.com/jukylin/esim/log"
	"net/http"
	"time"
)

type slowProxy struct {
	RoundTripWasCalled bool

	name string

	log log.Logger

	nextTransport http.RoundTripper
}

type slowProxyOption func(c *slowProxy)

type slowProxyOptions struct{}

func NewSlowProxy(logger log.Logger, name string) *slowProxy {
	slowProxy := &slowProxy{}

	if logger == nil {
		slowProxy.log = log.NewLogger()
	} else {
		slowProxy.log = logger
	}

	slowProxy.name = name

	return slowProxy
}

func (this *slowProxy) NextProxy(tripper interface{}) {
	this.nextTransport = tripper.(http.RoundTripper)
}

func (this *slowProxy) ProxyName() string {
	return this.name
}

// RoundTrip implements the RoundTripper interface.
func (this *slowProxy) RoundTrip(req *http.Request) (*http.Response, error) {
	if this.nextTransport == nil {
		this.nextTransport = http.DefaultTransport
	}
	time.Sleep(10 * time.Millisecond)
	resp, err := this.nextTransport.RoundTrip(req)

	return resp, err
}
