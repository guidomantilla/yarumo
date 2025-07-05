package comm

import (
	"fmt"
	"net"
	"net/http"
	"time"

	retry "github.com/avast/retry-go/v4"
	"github.com/rs/zerolog/log"
)

type httpClient struct {
	*http.Client
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
// The returned http.Client is equipped with a custom Transport that enables
// connection reuse and sets sane defaults for TCP dial, TLS handshake, and
// response header timeouts.
//
// Prefer this function over setting http.Client.Timeout when you need fine-grained
// control per request or want to avoid global timeout conflicts.
func NewHTTPClient(opts ...HttpOption) HTTPClient {
	options := NewHttpOptions(opts...)
	return &httpClient{
		Client: &http.Client{
			Timeout: options.Timeout,
			Transport: &HttpLoggingRoundTripper{
				MaxRetries: options.MaxRetries,
				Next: &http.Transport{
					TLSClientConfig:     options.TLSClientConfig,
					MaxIdleConns:        100,
					MaxIdleConnsPerHost: 10,
					IdleConnTimeout:     90 * time.Second,
					DialContext: (&net.Dialer{
						Timeout:   5 * time.Second,
						KeepAlive: 30 * time.Second,
					}).DialContext,
					TLSHandshakeTimeout:   5 * time.Second,
					ResponseHeaderTimeout: 10 * time.Second,
					ExpectContinueTimeout: 1 * time.Second,
				},
			},
		},
	}
}

func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	return c.Client.Do(req)
}

//

type HttpLoggingRoundTripper struct {
	MaxRetries uint
	Next       http.RoundTripper
}

func NewHttpLoggingRoundTripper(maxRetries uint, next http.RoundTripper) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	return &HttpLoggingRoundTripper{
		MaxRetries: maxRetries,
		Next:       next,
	}
}

func (lrt *HttpLoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	var reqBodyPreview string
	if req.Body != nil {
		body, buffer, err := ToReadNopCloser(req.Body)
		if err != nil {
			log.Error().Str("method", req.Method).Stringer("url", req.URL).Err(err).Msg("error reading request body")
			return nil, err
		}
		reqBodyPreview = string(buffer)
		req.Body = body
	}

	retryableCall := func() (*http.Response, error) {
		return lrt.Next.RoundTrip(req)
	}

	resp, err := retry.DoWithData(retryableCall, retry.Attempts(lrt.MaxRetries-1),
		retry.OnRetry(func(_ uint, err error) {
			log.Error().Str("method", req.Method).Stringer("url", req.URL).Str("requestBody", reqBodyPreview).Err(err).Msg("HTTP request failed")
		}),
	)

	if err != nil {
		log.Error().Str("method", req.Method).Stringer("url", req.URL).Str("requestBody", reqBodyPreview).Err(err).Msg(fmt.Sprintf("HTTP request failed after %d retries", lrt.MaxRetries))
		return nil, err
	}

	duration := time.Since(start)

	var respBodyPreview string
	if resp.Body != nil {
		body, buffer, err := ToReadNopCloser(resp.Body)
		if err != nil {
			log.Error().Str("method", req.Method).Stringer("url", req.URL).Err(err).Msg("error reading response body")
			return nil, err
		}
		respBodyPreview = string(buffer)
		resp.Body = body
	}

	log.Info().Str("method", req.Method).Stringer("url", req.URL).Int("status", resp.StatusCode).Dur("duration", duration).Str("requestBody", reqBodyPreview).Str("responseBody", respBodyPreview).Msg("HTTP request completed")
	return resp, nil
}
