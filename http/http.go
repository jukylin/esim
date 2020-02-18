package http

import (
	"context"
	"github.com/jukylin/esim/log"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"github.com/jukylin/esim/proxy"
)

type HttpClient struct {
	client *http.Client

	transports []func() interface{}

	logger log.Logger
}


type Option func(c *HttpClient)

type ClientOptions struct{}

func NewHttpClient(options ...Option) *HttpClient {

	httpClient := &HttpClient{
		transports: make([]func() interface{}, 0),
	}

	client := &http.Client{}
	httpClient.client = client

	for _, option := range options {
		option(httpClient)
	}

	if httpClient.transports == nil {
		httpClient.client.Transport = http.DefaultTransport
	}else{
		httpClient.client.Transport = proxy.NewProxyFactory().
			GetFirstInstance("http", http.DefaultTransport, httpClient.transports...).(http.RoundTripper)
	}

	if httpClient.client.Timeout <= 0 {
		httpClient.client.Timeout = 30 * time.Second
	}

	if httpClient.logger == nil {
		httpClient.logger = log.NewLogger()
	}

	return httpClient
}

func (ClientOptions) WithProxy(proxy ...func () interface{}) Option {
	return func(hc *HttpClient) {
		hc.transports = append(hc.transports, proxy...)
	}
}


func (ClientOptions) WithTimeOut(timeout time.Duration) Option {
	return func(hc *HttpClient) {
		hc.client.Timeout = timeout * time.Second
	}
}


func (ClientOptions) WithLogger(logger log.Logger) Option {
	return func(hc *HttpClient) {
		hc.logger = logger
	}
}


func (this *HttpClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := this.client.Do(req)
	return resp, err
}


func (this *HttpClient) Get(ctx context.Context, url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	resp, err = this.client.Do(req)
	return resp, err
}


func (this *HttpClient) Post(ctx context.Context, url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", contentType)
	return this.Do(ctx, req)
}


func (this *HttpClient) PostForm(ctx context.Context, url string, data url.Values) (resp *http.Response, err error) {
	return this.Post(ctx, url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}


func (this *HttpClient) Head(ctx context.Context, url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return this.Do(ctx, req)
}


func (this *HttpClient) CloseIdleConnections(ctx context.Context) {
	this.client.CloseIdleConnections()
}