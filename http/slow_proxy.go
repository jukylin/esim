package http

import (
	"net/http"
	"time"

	"github.com/jukylin/esim/log"
)

type slowProxy struct {
	RoundTripWasCalled bool

	name string

	log log.Logger

	nextTransport http.RoundTripper
}

func newSlowProxy(logger log.Logger, name string) *slowProxy {
	slowProxy := &slowProxy{}

	if logger == nil {
		slowProxy.log = log.NewLogger()
	} else {
		slowProxy.log = logger
	}

	slowProxy.name = name

	return slowProxy
}

func (sp *slowProxy) NextProxy(tripper interface{}) {
	sp.nextTransport = tripper.(http.RoundTripper)
}

func (sp *slowProxy) ProxyName() string {
	return sp.name
}

// RoundTrip implements the RoundTripper interface.
func (sp *slowProxy) RoundTrip(req *http.Request) (*http.Response, error) {
	if sp.nextTransport == nil {
		sp.nextTransport = http.DefaultTransport
	}
	time.Sleep(10 * time.Millisecond)
	resp, err := sp.nextTransport.RoundTrip(req)

	return resp, err
}
