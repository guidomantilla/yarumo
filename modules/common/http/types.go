package http

import "net/http"

var (
	_ Client = (*http.Client)(nil)
	_ Client = (*client)(nil)
	_ Client = (*MockClient)(nil)
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}
