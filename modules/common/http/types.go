package http

import "net/http"

var (
	_ Client = (*client)(nil)
	_ Client = (*http.Client)(nil)
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}
