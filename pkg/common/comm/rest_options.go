package comm

import "net/http"

type RestOptions struct {
	http    HTTPClient
	headers http.Header
}

func NewRestOptions(opts ...RestOption) *RestOptions {
	options := &RestOptions{
		http: http.DefaultClient,
		headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type RestOption func(opts *RestOptions)

func WithHTTPClient(client HTTPClient) RestOption {
	return func(opts *RestOptions) {
		opts.http = client
	}
}

func WithHeaders(headers http.Header) RestOption {
	return func(opts *RestOptions) {
		opts.headers = headers
	}
}
