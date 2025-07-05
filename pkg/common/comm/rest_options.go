package comm

import "net/http"

type RestOptions struct {
	http HTTPClient
}

func NewRestOptions(opts ...RestOption) *RestOptions {
	options := &RestOptions{
		http: http.DefaultClient,
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
