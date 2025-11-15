package http

import (
	"context"
	"net/http"
	"time"

	retry "github.com/avast/retry-go/v4"
	"golang.org/x/time/rate"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	"github.com/guidomantilla/yarumo/common/utils"
)

type client struct {
	http.Client
	attempts        uint
	retryIf         retry.RetryIfFunc
	retryHook       retry.OnRetryFunc
	limiter         *rate.Limiter
	retryOnResponse RetryOnResponseFunc
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
		attempts:        opts.attempts,
		retryIf:         opts.retryIf,
		retryHook:       opts.retryHook,
		limiter:         rate.NewLimiter(opts.limiterRate, int(opts.limiterBurst)), //nolint:gosec // disable G115
		retryOnResponse: opts.retryOnResponse,
	}
}

// RetrierEnabled returns true if retries are enabled.
// A retry is enabled if attempts > 1.
func (c *client) RetrierEnabled() bool {
	return c.attempts > 1
}

// LimiterEnabled returns true if the limiter is enabled.
// A limiter is enabled if its rate is finite and its burst is > 0.
func (c *client) LimiterEnabled() bool {
	return utils.NotEmpty(c.limiter) && utils.NotEqual(c.limiter.Limit(), rate.Inf)
}

// Do execute the request. If the limiter is active, it may wait for a token before performing the request.
// It may retry the request if configured to do so through Options. It returns the first successful response.
// The caller must close res.Body when err == nil.
func (c *client) Do(req *http.Request) (*http.Response, error) {

	// Safety check: if there is a body, and we don't have GetBody, we cannot retry safely.
	if utils.NotEmpty(req.Body) && utils.Empty(req.GetBody) {
		return nil, ErrDo(ErrHttpNonReplayableBody, nil)
	}

	retryableCall := func() (*http.Response, error) {
		// Clone the base request. Clone makes a shallow copy of fields (including GetBody),
		// but it does NOT recreate the body: we do that ourselves.
		clonedReq := req.Clone(req.Context())
		if utils.NotEmpty(req.Body, req.GetBody) {
			rc, err := req.GetBody()
			if err != nil {
				return nil, ErrDo(ErrHttpGetBodyFailed, err)
			}
			clonedReq.Body = rc
		}

		err := c.waitForLimiter(clonedReq.Context())
		if err != nil {
			return nil, ErrDo(err)
		}

		res, err := c.Client.Do(clonedReq)
		if err != nil {
			// When an error is returned, the standard net/http client ignores any response value.
			// Just wrap and return the error.
			return nil, ErrDo(ErrHttpRequestFailed, err)
		}

		if c.retryOnResponse(res) {
			_ = res.Body.Close()
			return nil, ErrDo(&StatusCodeError{StatusCode: res.StatusCode})
		}

		return res, nil
	}

	return retry.DoWithData(retryableCall,
		retry.Attempts(c.attempts), retry.RetryIf(c.retryIf), retry.OnRetry(c.retryHook))
}

// waitForLimiter blocks until a token is available from the limiter.
// It returns an error if the limiter is disabled or if the context expires before a token is available.
func (c *client) waitForLimiter(ctx context.Context) error {

	// Only wait on the limiter when it is effectively enabled.
	// Semantics: rate.Inf means limiter is disabled.
	if !c.LimiterEnabled() {
		return nil
	}

	// Bound the limiter wait by the effective deadline, which is the minimum between req.Context() deadline and the client timeout.
	waitCtx := ctx
	if c.Timeout > 0 {
		clientDeadline := time.Now().Add(c.Timeout)
		deadline, ok := waitCtx.Deadline()
		if !ok || clientDeadline.Before(deadline) {
			var cancel context.CancelFunc
			waitCtx, cancel = context.WithDeadline(waitCtx, clientDeadline)
			defer cancel()
		}
	}

	if err := c.limiter.Wait(waitCtx); err != nil {
		return cerrs.Wrap(ErrRateLimiterExceeded, err)
	}

	return nil
}
