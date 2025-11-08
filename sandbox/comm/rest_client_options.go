package comm

import "net/http"

type RestClientOptions struct {
	http             HTTPClient
	headers          http.Header
	statusCodeErrors []int
	statusCodeOK     []int
}

func NewRestClientOptions(opts ...RestClientOption) *RestClientOptions {
	options := &RestClientOptions{
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

type RestClientOption func(opts *RestClientOptions)

func WithHTTPClient(client HTTPClient) RestClientOption {
	return func(opts *RestClientOptions) {
		opts.http = client
	}
}

func WithHeaders(headers http.Header) RestClientOption {
	return func(opts *RestClientOptions) {
		opts.headers = headers
	}
}

func WithStatusCodeErrors(codes []int) RestClientOption {
	return func(opts *RestClientOptions) {
		opts.statusCodeErrors = codes
	}
}

func WithStatusCodeOK(codes []int) RestClientOption {
	return func(opts *RestClientOptions) {
		opts.statusCodeOK = codes
	}
}
