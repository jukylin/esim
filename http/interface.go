package http


import (
	"context"
	"net/http"
	"io"
	"net/url"
)

type HttpClient interface {

	Do(context.Context, *http.Request) (*http.Response, error)

	Get(context.Context, string) (*http.Response, error)

	Post(context.Context, string, string, io.Reader) (*http.Response, error)

	PostForm(context.Context, string, url.Values) (*http.Response, error)

	Head(context.Context, string) (*http.Response, error)
}