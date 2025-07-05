package comm

import "net/http"

type RestOptions struct {
	http             HTTPClient
	headers          http.Header
	statusCodeErrors []int
	statusCodeOK     []int
}

func NewRestOptions(opts ...RestOption) *RestOptions {
	options := &RestOptions{
		http: http.DefaultClient,
		headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
		statusCodeErrors: []int{
			http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound,
			http.StatusConflict, http.StatusUnprocessableEntity, http.StatusTooManyRequests,
			http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout,
		},
		statusCodeOK: []int{http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent, http.StatusPartialContent},
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

func WithStatusCodeErrors(codes []int) RestOption {
	return func(opts *RestOptions) {
		opts.statusCodeErrors = codes
	}
}

func WithStatusCodeOK(codes []int) RestOption {
	return func(opts *RestOptions) {
		opts.statusCodeOK = codes
	}
}
