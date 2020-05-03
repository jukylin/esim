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

func NewSpyProxy(logger log.Logger, name string) *spyProxy {
	spyProxy := &spyProxy{}

	if logger == nil {
		spyProxy.log = log.NewLogger()
	} else {
		spyProxy.log = logger
	}

	spyProxy.name = name

	return spyProxy
}

func (sp *spyProxy) NextProxy(tripper interface{}) {
	sp.nextTransport = tripper.(http.RoundTripper)
}

func (sp *spyProxy) ProxyName() string {
	return sp.name
}

// RoundTrip implements the RoundTripper interface.
func (sp *spyProxy) RoundTrip(req *http.Request) (*http.Response, error) {
	if sp.nextTransport == nil {
		sp.nextTransport = http.DefaultTransport
	}

	sp.RoundTripWasCalled = true
	sp.log.Infof("%s was called", sp.name)
	resp, err := sp.nextTransport.RoundTrip(req)

	return resp, err
}
