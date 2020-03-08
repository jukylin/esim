package http

import (
	"github.com/jukylin/esim/log"
	"net/http"
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

func (this *stubsProxy) NextProxy(tripper interface{}) {
	this.nextTransport = tripper.(http.RoundTripper)
}

func (this *stubsProxy) ProxyName() string {
	return this.name
}

// RoundTrip implements the RoundTripper interface.
func (this *stubsProxy) RoundTrip(req *http.Request) (*http.Response, error) {
	return this.respFunc(req), nil
}
