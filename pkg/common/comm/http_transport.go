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
	logger := log.With().Str("stage", "runtime").Str("component", "http-transport").Str("method", req.Method).Stringer("url", req.URL).Logger()
	start := time.Now()

	reqBody := []byte("{}")
	reqHeaders := MustJsonMarshalSanitized(req.Header)
	if req.Body != nil {
		body, buffer, err := ToReadNopCloser(req.Body)
		if err != nil {
			logger.Error().Err(err).Msg("error reading request body")
			logger.Trace().RawJSON("req-headers", reqHeaders).Err(err).Msg("error reading request body")
			return nil, fmt.Errorf("error reading request body: %w", err)
		}
		reqBody = buffer
		req.Body = body
	}

	retryableCall := func() (*http.Response, error) {
		return transport.Next.RoundTrip(req)
	}

	res, err := retry.DoWithData(retryableCall, retry.Attempts(transport.MaxRetries-1),
		retry.OnRetry(func(_ uint, err error) {
			logger.Error().Err(err).Msg("HTTP request failed")
		}),
	)

	if err != nil {
		logger.Error().Err(err).Msg(fmt.Sprintf("HTTP request failed after %d retries", transport.MaxRetries))
		logger.Trace().Int("status", res.StatusCode).
			RawJSON("req-headers", reqHeaders).Func(AppendBody(req.Header, "req-body", reqBody)).
			Err(err).Msg(fmt.Sprintf("HTTP request failed after %d retries", transport.MaxRetries))
		return nil, fmt.Errorf("HTTP request failed after %d retries: %w", transport.MaxRetries, err)
	}

	resBody := []byte("{}")
	resHeaders := MustJsonMarshalSanitized(res.Header)
	if res.Body != nil {
		body, buffer, err := ToReadNopCloser(res.Body)
		if err != nil {
			logger.Error().Err(err).Msg("error reading response body")
			logger.Trace().Int("status", res.StatusCode).
				RawJSON("req-headers", reqHeaders).Func(AppendBody(req.Header, "req-body", reqBody)).
				RawJSON("res-headers", resHeaders).Func(AppendBody(res.Header, "res-body", resBody)).
				Err(err).Msg("error reading response body")
			return nil, fmt.Errorf("error reading response body: %w", err)
		}
		resBody = buffer
		res.Body = body
	}

	duration := time.Since(start)
	logger.Trace().Int("status", res.StatusCode).Dur("duration", duration).
		RawJSON("req-headers", reqHeaders).Func(AppendBody(req.Header, "req-body", reqBody)).
		RawJSON("res-headers", resHeaders).Func(AppendBody(res.Header, "res-body", resBody)).
		Msg("HTTP request completed")
	return res, nil
}
