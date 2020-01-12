package http

import (
	"net/http"
	"io/ioutil"
	"bytes"
	"github.com/jukylin/esim/log"
)

type stubsProxy struct {

	name string

	log log.Logger

	nextTransport http.RoundTripper
}

type stubsProxyOption func(c *stubsProxy)

type stubsProxyOptions struct{}

func NewStubsProxy(logger log.Logger, name string) *stubsProxy {
	stubsProxy := &stubsProxy{}

	if logger == nil{
		stubsProxy.log = log.NewLogger()
	}else{
		stubsProxy.log = logger
	}

	stubsProxy.name = name

	return stubsProxy
}

func (this *stubsProxy) NextProxy(tripper interface{})  {
	this.nextTransport = tripper.(http.RoundTripper)
}

func (this *stubsProxy) ProxyName() string {
	return this.name
}

// RoundTrip implements the RoundTripper interface.
func (this *stubsProxy) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := &http.Response{}
	if req.URL.String() == "127.0.0.1"{
		resp.StatusCode = 200
	}else if req.URL.String() == "127.0.0.2"{
		resp.StatusCode = 300
	}

	resp.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))

	return resp, nil
}