package comm

import (
	"context"
	"net/http"
)

var (
	_ HTTPClient        = (*httpClient)(nil)
	_ HTTPClient        = (*http.Client)(nil)
	_ http.RoundTripper = (*HttpLoggingRoundTripper)(nil)
	_ RESTCallFn[any]   = RESTCall[any]
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type RESTCallFn[T any] func(ctx context.Context, method string, url string, body any, headers http.Header, opts ...RestOption) (*RESTResponse[T], error)
