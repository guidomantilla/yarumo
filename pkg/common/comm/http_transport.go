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
			log.Trace().Str("method", req.Method).Stringer("url", req.URL).
				Stringer("req-headers", HttpHeader(req.Header)).Str("req-body", reqBodyPreview).
				Err(err).Msg("error reading request body")
			return nil, err
		}
		reqBodyPreview = string(buffer)
		req.Body = body
	}

	retryableCall := func() (*http.Response, error) {
		return transport.Next.RoundTrip(req)
	}

	res, err := retry.DoWithData(retryableCall, retry.Attempts(transport.MaxRetries-1),
		retry.OnRetry(func(_ uint, err error) {
			log.Error().Str("method", req.Method).Stringer("url", req.URL).Err(err).Msg("HTTP request failed")
		}),
	)

	if err != nil {
		log.Error().Str("method", req.Method).Stringer("url", req.URL).Err(err).Msg(fmt.Sprintf("HTTP request failed after %d retries", transport.MaxRetries))
		log.Trace().Str("method", req.Method).Stringer("url", req.URL).Int("status", res.StatusCode).
			Stringer("req-headers", HttpHeader(req.Header)).Str("req-body", reqBodyPreview).
			Err(err).Msg(fmt.Sprintf("HTTP request failed after %d retries", transport.MaxRetries))
		return nil, err
	}

	duration := time.Since(start)

	var resBodyPreview string
	if res.Body != nil {
		body, buffer, err := ToReadNopCloser(res.Body)
		if err != nil {
			log.Error().Str("method", req.Method).Stringer("url", req.URL).Err(err).Msg("error reading response body")
			log.Trace().Str("method", req.Method).Stringer("url", req.URL).Int("status", res.StatusCode).
				Stringer("req-headers", HttpHeader(req.Header)).Str("req-body", reqBodyPreview).
				Err(err).Msg("error reading response body")
			return nil, err
		}
		resBodyPreview = string(buffer)
		res.Body = body
	}

	log.Trace().Str("method", req.Method).Stringer("url", req.URL).Int("status", res.StatusCode).Dur("duration", duration).
		Stringer("req-headers", HttpHeader(req.Header)).Str("req-body", reqBodyPreview).
		Stringer("res-headers", HttpHeader(res.Header)).Str("res-body", resBodyPreview).
		Msg("HTTP request completed")
	return res, nil
}
