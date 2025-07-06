package comm

import (
	"fmt"
	"net/http"
	"time"

	retry "github.com/avast/retry-go/v4"
	"github.com/rs/zerolog/log"
)

type HttpTransport struct {
	MaxRetries uint
	Next       http.RoundTripper
}

func NewHttpTransport(opts ...HttpTransportOption) *HttpTransport {
	options := NewHttpTransportOptions(opts...)
	return &HttpTransport{
		MaxRetries: options.MaxRetries,
		Next: &http.Transport{
			TLSClientConfig:       options.TLSClientConfig,
			MaxIdleConns:          options.MaxIdleConns,
			MaxIdleConnsPerHost:   options.MaxIdleConnsPerHost,
			IdleConnTimeout:       options.IdleConnTimeout,
			DialContext:           options.DialContext,
			TLSHandshakeTimeout:   options.TLSHandshakeTimeout,
			ResponseHeaderTimeout: options.ResponseHeaderTimeout,
			ExpectContinueTimeout: options.ExpectContinueTimeout,
		},
	}
}

func (transport *HttpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
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
		return transport.Next.RoundTrip(req)
	}

	resp, err := retry.DoWithData(retryableCall, retry.Attempts(transport.MaxRetries-1),
		retry.OnRetry(func(_ uint, err error) {
			log.Error().Str("method", req.Method).Stringer("url", req.URL).Str("requestBody", reqBodyPreview).Err(err).Msg("HTTP request failed")
		}),
	)

	if err != nil {
		log.Error().Str("method", req.Method).Stringer("url", req.URL).Str("requestBody", reqBodyPreview).Err(err).Msg(fmt.Sprintf("HTTP request failed after %d retries", transport.MaxRetries))
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
