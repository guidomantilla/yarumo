package retry

import (
	"errors"
	"net/http"
)

// NoopRetryOnResponse always returns false; disables response-based retries.
func NoopRetryOnResponse(_ *http.Response) bool {
	return false
}

// NoopRetryIf always returns false; disables error-based retries.
func NoopRetryIf(_ error) bool {
	return false
}

// NoopRetryHook is a no-op retry hook.
func NoopRetryHook(_ uint, _ error) {}

// RetryOn5xxAnd429 returns true for 5xx server errors and 429 throttling.
// Other status codes are not retried.
func RetryOn5xxAnd429(res *http.Response) bool {
	if res == nil {
		return false
	}

	if res.StatusCode == http.StatusTooManyRequests {
		return true
	}

	return res.StatusCode >= 500 && res.StatusCode < 600
}

// RetryIfHttpError reports whether err is a *StatusCodeError synthesized
// by the retry transport when a RetryOnResponseFn returns true. Use it
// (or errors.As) to decide whether to retry on response-driven failures.
func RetryIfHttpError(err error) bool {
	var sce *StatusCodeError
	return errors.As(err, &sce)
}
