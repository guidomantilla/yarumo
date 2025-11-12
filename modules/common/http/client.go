package http

import (
	"net/http"

	retry "github.com/avast/retry-go/v4"
	"github.com/guidomantilla/yarumo/common/utils"
	"golang.org/x/time/rate"
)

type client struct {
	http.Client
	attempts  uint
	retryIf   retry.RetryIfFunc
	retryHook retry.OnRetryFunc
	limiter   *rate.Limiter
}

// NewClient creates a client compatible with *http.Client that can
// apply rate limiting and retries according to provided Options.
//
// If limiterRate == rate.Inf, the limiter is effectively disabled.
// If limiterRate is finite and limiterBurst <= 0, it is normalized to burst=1.
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

// Do execute the request. If the limiter is active, it may wait for
// a token before performing the request. It may retry the request if
// configured to do so through Options. It returns the first successful
// response. The caller must close res.Body when err == nil.
func (c *client) Do(req *http.Request) (*http.Response, error) {
	retryableCall := func() (*http.Response, error) {

		// Only wait on the limiter when it is effectively enabled.
		// Semantics: rate.Inf means limiter is disabled.
		if c.RateLimiterEnabled() {
			err := c.limiter.Wait(req.Context())
			if err != nil {
				return nil, ErrDoCall(ErrRateLimiterExceeded, err)
			}
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

func (c *client) RateLimiterEnabled() bool {
	return utils.NotEmpty(c.limiter) && utils.NotEqual(c.limiter.Limit(), rate.Inf)
}

func (c *client) RoundTrip(req *http.Request) (*http.Response, error) {
	return c.Transport.RoundTrip(req)
}
