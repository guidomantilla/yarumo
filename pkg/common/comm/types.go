package comm

import "net/http"

var (
	_ HTTPClient        = (*httpClient)(nil)
	_ HTTPClient        = (*http.Client)(nil)
	_ http.RoundTripper = (*LoggingRoundTripper)(nil)
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
