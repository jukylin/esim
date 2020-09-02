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

// GlobalStub is test double and it is used when we cannot or don’t want to involve real server.
// Instead of the real server, we introduced a stub and defined what data should be returned.
// Example:
//	func(request *http.Request) *http.Response {
//		resp := &http.Response{}
//		if request.URL.String() == "127.0.0.1" {
//			resp.StatusCode = http.StatusOK
//			resp.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))
//		}
//		return resp
//	}
var GlobalStub func(*http.Request) *http.Response

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

	if Client.logger == nil {
		Client.logger = log.NewLogger()
	}

	if GlobalStub != nil {
		// It is the last transports.
		Client.transports = append(Client.transports, newGlobalStub(GlobalStub, Client.logger))
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

func newGlobalStub(stubFunc func(*http.Request) *http.Response, logger log.Logger) func() interface{} {
	return 	func() interface{} {
		stubsProxyOptions := StubsProxyOptions{}
		stubsProxy := NewStubsProxy(
			stubsProxyOptions.WithRespFunc(stubFunc),
			stubsProxyOptions.WithName("global stub"),
			stubsProxyOptions.WithLogger(logger),
		)

		return stubsProxy
	}
}