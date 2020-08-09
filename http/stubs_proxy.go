package http

import (
	"net/http"

	"github.com/jukylin/esim/log"
)

type StubsProxy struct {
	name string

	logger log.Logger

	nextTransport http.RoundTripper

	respFunc func(r *http.Request) *http.Response
}

type StubsProxyOption func(c *StubsProxy)

type StubsProxyOptions struct{}

type RespFunc func(*http.Request) *http.Response

func NewStubsProxy(options ...StubsProxyOption) *StubsProxy {
	stubsProxy := &StubsProxy{}

	for _, option := range options {
		option(stubsProxy)
	}

	if stubsProxy.logger == nil {
		stubsProxy.logger = log.NewLogger()
	}

	return stubsProxy
}

func (StubsProxyOptions) WithRespFunc(respFunc RespFunc) StubsProxyOption {
	return func(s *StubsProxy) {
		s.respFunc = respFunc
	}
}

func (StubsProxyOptions) WithName(name string) StubsProxyOption {
	return func(s *StubsProxy) {
		s.name = name
	}
}

func (StubsProxyOptions) WithLogger(logger log.Logger) StubsProxyOption {
	return func(s *StubsProxy) {
		s.logger = logger
	}
}

func (sp *StubsProxy) NextProxy(tripper interface{}) {
	sp.nextTransport = tripper.(http.RoundTripper)
}

func (sp *StubsProxy) ProxyName() string {
	return sp.name
}

// RoundTrip implements the RoundTripper interface.
func (sp *StubsProxy) RoundTrip(req *http.Request) (*http.Response, error) {
	sp.logger.Infoc(req.Context(), "Call %s", sp.name)
	return sp.respFunc(req), nil
}
