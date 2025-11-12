package comm

import (
	"context"
	"net/http"
)

var (
	_ HTTPClient = (*http.Client)(nil)
	_ RESTClient = (*restClient)(nil)
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type RESTClient interface {
	Call(ctx context.Context, method string, path string, body any) (*RESTClientResponse, error)
}
