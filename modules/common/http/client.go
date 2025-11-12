package http

import (
    "context"
    "net/http"
    "time"

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
        if c.LimiterEnabled() {
            // Bound the limiter wait by the effective deadline which is the
            // minimum between req.Context() deadline and the client timeout.
            waitCtx := req.Context()
            if c.Timeout > 0 {
                // Compute client deadline as now + client timeout
                clientDeadline := time.Now().Add(c.Timeout)
                if dl, ok := waitCtx.Deadline(); ok {
                    // Use the earlier deadline between context and client
                    if clientDeadline.Before(dl) {
                        var cancel context.CancelFunc
                        waitCtx, cancel = context.WithDeadline(waitCtx, clientDeadline)
                        defer cancel()
                    }
                } else {
                    var cancel context.CancelFunc
                    waitCtx, cancel = context.WithDeadline(waitCtx, clientDeadline)
                    defer cancel()
                }
            }

            err := c.limiter.Wait(waitCtx)
            if err != nil {
                return nil, ErrDoCall(ErrRateLimiterExceeded, err)
            }
        }

		res, err := c.Client.Do(req)
		if err != nil {
			// Defensive: if an implementation returns both a response and an error,
			// ensure the body is closed to avoid leaking connections before retrying
			// or returning to the caller.
			if res != nil && res.Body != nil {
				_ = res.Body.Close()
			}
			return nil, ErrDoCall(ErrHttpRequestFailed, err)
		}

		return res, nil
	}

	return retry.DoWithData(retryableCall,
		retry.Attempts(c.attempts), retry.RetryIf(c.retryIf), retry.OnRetry(c.retryHook))
}

func (c *client) LimiterEnabled() bool {
	return utils.NotEmpty(c.limiter) && utils.NotEqual(c.limiter.Limit(), rate.Inf)
}
