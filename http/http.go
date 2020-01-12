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

type httpClient struct {
	*http.Client

	transports []func() interface{}

	log log.Logger
}

type Option func(c *httpClient)

type ClientOptions struct{}

func NewHttpClient(options ...Option) HttpClient {

	httpClient := &httpClient{
		transports: make([]func() interface{}, 0),
	}

	client := &http.Client{}
	httpClient.Client = client

	for _, option := range options {
		option(httpClient)
	}

	if httpClient.transports == nil {
		httpClient.Client.Transport = http.DefaultTransport
	}else{
		httpClient.Client.Transport = proxy.NewProxyFactory().
			GetFirstInstance("http", http.DefaultTransport, httpClient.transports...).(http.RoundTripper)
	}

	if httpClient.Timeout <= 0 {
		httpClient.Client.Timeout = 3 * time.Second
	}

	if httpClient.log == nil {
		httpClient.log = log.NewLogger()
	}

	return httpClient
}

func (ClientOptions) WithProxy(proxy ...func () interface{}) Option {
	return func(hc *httpClient) {
		hc.transports = append(hc.transports, proxy...)
	}
}

func (ClientOptions) WithTimeOut(timeout time.Duration) Option {
	return func(hc *httpClient) {
		hc.Client.Timeout = timeout * time.Second
	}
}

func (ClientOptions) WithLogger(log log.Logger) Option {
	return func(hc *httpClient) {
		hc.log = log
	}
}

func (this *httpClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := this.Client.Do(req)
	return resp, err
}

func (this *httpClient) Get(ctx context.Context, url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	resp, err = this.Client.Do(req)
	return resp, err
}

func (this *httpClient) Post(ctx context.Context, url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", contentType)
	return this.Do(ctx, req)
}

func (this *httpClient) PostForm(ctx context.Context, url string, data url.Values) (resp *http.Response, err error) {
	return this.Post(ctx, url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (this *httpClient) Head(ctx context.Context, url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return this.Do(ctx, req)
}