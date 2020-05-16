package http

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/proxy"
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
			GetFirstInstance("http", http.DefaultTransport,
				Client.transports...).(http.RoundTripper)
	}

	if Client.client.Timeout <= 0 {
		Client.client.Timeout = 30 * time.Second
	}

	if Client.logger == nil {
		Client.logger = log.NewLogger()
	}

	return Client
}

func (ClientOptions) WithProxy(proxys ...func() interface{}) Option {
	return func(hc *Client) {
		hc.transports = append(hc.transports, proxys...)
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

func (c *Client) Get(ctx context.Context, addr string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	resp, err = c.client.Do(req)
	return resp, err
}

func (c *Client) Post(ctx context.Context, addr, contentType string,
	body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", addr, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", contentType)
	return c.Do(ctx, req)
}

func (c *Client) PostForm(ctx context.Context, addr string,
	data url.Values) (resp *http.Response, err error) {
	return c.Post(ctx, addr, "application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()))
}

func (c *Client) Head(ctx context.Context, addr string) (resp *http.Response, err error) {
	req, err := http.NewRequest("HEAD", addr, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return c.Do(ctx, req)
}

func (c *Client) CloseIdleConnections(ctx context.Context) {
	c.client.CloseIdleConnections()
}
