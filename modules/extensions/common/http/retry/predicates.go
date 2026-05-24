package retry

import (
	"errors"
	"net/http"
)

// NoopRetryOnResponse always returns false; disables response-based retries.
// This is the default value of WithRetryOnResponse.
func NoopRetryOnResponse(_ *http.Response) bool {
	return false
}

// RetryOn5xxAnd429 returns true for 5xx server errors and 429 throttling.
// Other status codes are not retried. Pass it to WithRetryOnResponse to
// opt into HTTP-status-driven retries.
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
// with the underlying retry.Retry's WithRetryIf to opt into
// response-driven retries (alongside any predicates for transport
// errors).
func RetryIfHttpError(err error) bool {
	var sce *StatusCodeError
	return errors.As(err, &sce)
}
