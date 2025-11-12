package http

import (
	"net/http"

	retry "github.com/avast/retry-go/v4"
	"golang.org/x/time/rate"
)

type client struct {
	http.Client
	attempts  uint
	retryIf   retry.RetryIfFunc
	retryHook retry.OnRetryFunc
	limiter   *rate.Limiter
}

func NewClient(options ...Option) Client {
	opts := NewOptions(options...)
	return &client{
		Client: http.Client{
			Timeout:   opts.timeout,
			Transport: opts.transport,
		},
		attempts:  opts.attempts,
		retryIf:   opts.retryIf,
		retryHook: opts.retryHook,
		limiter:   rate.NewLimiter(opts.limiterRate, opts.limiterBurst),
	}
}

func (c *client) Do(req *http.Request) (*http.Response, error) {
	retryableCall := func() (*http.Response, error) {

		err := c.limiter.Wait(req.Context())
		if err != nil {
			return nil, ErrDoCall(ErrRateLimiterExceeded, err)
		}

		res, err := c.Client.Do(req)
		if err != nil {
			return nil, ErrDoCall(ErrHttpRequestFailed, err)
		}

		return res, nil
	}

	return retry.DoWithData(retryableCall,
		retry.Attempts(c.attempts), retry.RetryIf(c.retryIf), retry.OnRetry(c.retryHook))
}

func (c *client) RoundTrip(req *http.Request) (*http.Response, error) {
	return c.Transport.RoundTrip(req)
}
