package comm

import (
	"net/http"
)

type httpClient struct {
	http.Client
}

// NewHTTPClient returns a configured *http.Client
// with a secure and efficient Transport, but without a global timeout unless specified in options.
//
// This is intended for applications where request timeouts
// are managed externally using context.Context. It allows for more
// granular and flexible timeout control—especially useful in microservices
// or distributed systems where request lifetimes are propagated via context.
//
// IMPORTANT: If no context with timeout or cancellation is passed to the request,
// the HTTP call may block indefinitely. Always use a context like:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	req, _ := http.NewRequestWithContext(ctx, "GET", "https://...", nil)
//
// The returned http.Client is equipped with custom Transport that enables
// connection reuse and sets sane defaults for TCP dial, TLS handshake, and
// response header timeouts.
//
// Prefer this function over setting http.Client.Timeout when you need fine-grained
// control per request or want to avoid global timeout conflicts.
func NewHTTPClient(opts ...HttpClientOption) HTTPClient {
	options := NewHttpClientOptions(opts...)
	return &httpClient{
		Client: http.Client{
			Timeout:   options.timeout,
			Transport: options.transport,
		},
	}
}

func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	return c.Client.Do(req)
}
