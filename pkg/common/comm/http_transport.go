package comm

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	retry "github.com/avast/retry-go/v4"
	"github.com/rs/zerolog/log"
	"github.com/sony/gobreaker"

	"github.com/guidomantilla/yarumo/pkg/common/assert"
	"github.com/guidomantilla/yarumo/pkg/common/resilience"
)

type HttpTransport struct {
	MaxRetries             uint
	RateLimiterRegistry    *resilience.RateLimiterRegistry
	CircuitBreakerRegistry *resilience.CircuitBreakerRegistry
	Next                   http.RoundTripper
}

func NewHttpTransport(rateLimiterRegistry *resilience.RateLimiterRegistry, circuitBreakerRegistry *resilience.CircuitBreakerRegistry, opts ...HttpTransportOption) *HttpTransport {
	assert.NotEmpty(rateLimiterRegistry, fmt.Sprintf("%s - error creating: rateLimiterRegistry is empty", "http-transport"))
	assert.NotEmpty(circuitBreakerRegistry, fmt.Sprintf("%s - error creating: circuitBreakerRegistry is empty", "http-transport"))
	options := NewHttpTransportOptions(opts...)
	return &HttpTransport{
		MaxRetries:             options.MaxRetries,
		RateLimiterRegistry:    rateLimiterRegistry,
		CircuitBreakerRegistry: circuitBreakerRegistry,
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
			err = fmt.Errorf("error reading request body: %w", err)
			logger.Error().Err(err).Msg("error reading request body")
			logger.Trace().RawJSON("req-headers", reqHeaders).Err(err).Msg("error reading request body")
			return nil, err
		}
		reqBody = buffer
		req.Body = body
	}

	retryableCall := func() (*http.Response, error) {
		limiter := transport.RateLimiterRegistry.Get(fmt.Sprintf("http-transport-rate-limiter-%s", req.URL.Host))
		err := limiter.Wait(req.Context())
		if err != nil {
			return nil, fmt.Errorf("rate limit exceeded: %w", err)
		}

		breaker := transport.CircuitBreakerRegistry.Get(fmt.Sprintf("http-transport-circuit-breaker-%s", req.URL.Host))
		res, err := breaker.Execute(func() (any, error) {
			return transport.Next.RoundTrip(req)
		})
		if err != nil {
			return nil, fmt.Errorf("circuit breaker open or request failed: %w", err)
		}

		httpRes, ok := res.(*http.Response)
		if !ok || httpRes == nil {
			return nil, fmt.Errorf("unexpected result from breaker: %T", res)
		}

		return httpRes, nil
	}

	res, err := retry.DoWithData(retryableCall, retry.Attempts(transport.MaxRetries-1),
		retry.RetryIf(func(err error) bool {
			return !errors.Is(err, gobreaker.ErrOpenState)
		}),
		retry.OnRetry(func(_ uint, err error) {
			logger.Error().Err(err).Msg("HTTP request failed")
		}),
	)

	if err != nil {
		err = fmt.Errorf("HTTP request failed after %d retries: %w", transport.MaxRetries, err)
		logger.Error().Err(err).Msg(fmt.Sprintf("HTTP request failed after %d retries", transport.MaxRetries))
		logger.Trace().Int("status", res.StatusCode).
			RawJSON("req-headers", reqHeaders).Func(AppendBody(req.Header, "req-body", reqBody)).
			Err(err).Msg(fmt.Sprintf("HTTP request failed after %d retries", transport.MaxRetries))
		return nil, err
	}

	resBody := []byte("{}")
	resHeaders := MustJsonMarshalSanitized(res.Header)
	if res.Body != nil {
		body, buffer, err := ToReadNopCloser(res.Body)
		if err != nil {
			err = fmt.Errorf("error reading response body: %w", err)
			logger.Error().Err(err).Msg("error reading response body")
			logger.Trace().Int("status", res.StatusCode).
				RawJSON("req-headers", reqHeaders).Func(AppendBody(req.Header, "req-body", reqBody)).
				RawJSON("res-headers", resHeaders).Func(AppendBody(res.Header, "res-body", resBody)).
				Err(err).Msg("error reading response body")
			return nil, err
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
