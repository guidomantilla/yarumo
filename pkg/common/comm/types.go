package comm

import (
	"context"
	"net/http"
)

var (
	_ HTTPClient        = (*httpClient)(nil)
	_ HTTPClient        = (*http.Client)(nil)
	_ http.RoundTripper = (*HttpTransport)(nil)
	_ RESTClient        = (*restClient)(nil)
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type RESTClient interface {
	Call(ctx context.Context, method string, path string, body any) (*RESTResponse, error)
}
