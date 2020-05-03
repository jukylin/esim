package http

import (
	"net/http"

	"github.com/jukylin/esim/log"
)

type stubsProxy struct {
	name string

	logger log.Logger

	nextTransport http.RoundTripper

	respFunc func(r *http.Request) *http.Response
}

type StubsProxyOption func(c *stubsProxy)

type StubsProxyOptions struct{}

func NewStubsProxy(options ...StubsProxyOption) *stubsProxy {
	stubsProxy := &stubsProxy{}

	for _, option := range options {
		option(stubsProxy)
	}

	return stubsProxy
}

func (StubsProxyOptions) WithRespFunc(respFunc func(*http.Request) *http.Response) StubsProxyOption {
	return func(s *stubsProxy) {
		s.respFunc = respFunc
	}
}

func (StubsProxyOptions) WithName(name string) StubsProxyOption {
	return func(s *stubsProxy) {
		s.name = name
	}
}

func (StubsProxyOptions) WithLogger(logger log.Logger) StubsProxyOption {
	return func(s *stubsProxy) {
		s.logger = logger
	}
}

func (sp *stubsProxy) NextProxy(tripper interface{}) {
	sp.nextTransport = tripper.(http.RoundTripper)
}

func (sp *stubsProxy) ProxyName() string {
	return sp.name
}

// RoundTrip implements the RoundTripper interface.
func (sp *stubsProxy) RoundTrip(req *http.Request) (*http.Response, error) {
	return sp.respFunc(req), nil
}
