package http

import (
	"context"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/proxy"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	client *http.Client

	transports []func() interface{}

	logger log.Logger
}

type Option func(c *Client)

type ClientOptions struct{}

func NewClient(options ...Option) *Client {

	Client := &Client{
		transports: make([]func() interface{}, 0),
	}

	client := &http.Client{}
	Client.client = client

	for _, option := range options {
		option(Client)
	}

	if Client.transports == nil {
		Client.client.Transport = http.DefaultTransport
	} else {
		Client.client.Transport = proxy.NewProxyFactory().
			GetFirstInstance("http", http.DefaultTransport, Client.transports...).(http.RoundTripper)
	}

	if Client.client.Timeout <= 0 {
		Client.client.Timeout = 30 * time.Second
	}

	if Client.logger == nil {
		Client.logger = log.NewLogger()
	}

	return Client
}

func (ClientOptions) WithProxy(proxy ...func() interface{}) Option {
	return func(hc *Client) {
		hc.transports = append(hc.transports, proxy...)
	}
}

func (ClientOptions) WithTimeOut(timeout time.Duration) Option {
	return func(hc *Client) {
		hc.client.Timeout = timeout * time.Second
	}
}

func (ClientOptions) WithLogger(logger log.Logger) Option {
	return func(hc *Client) {
		hc.logger = logger
	}
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	return resp, err
}

func (c *Client) Get(ctx context.Context, url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	resp, err = c.client.Do(req)
	return resp, err
}

func (c *Client) Post(ctx context.Context, url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", contentType)
	return c.Do(ctx, req)
}

func (c *Client) PostForm(ctx context.Context, url string, data url.Values) (resp *http.Response, err error) {
	return c.Post(ctx, url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (c *Client) Head(ctx context.Context, url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return c.Do(ctx, req)
}

func (c *Client) CloseIdleConnections(ctx context.Context) {
	c.client.CloseIdleConnections()
}
