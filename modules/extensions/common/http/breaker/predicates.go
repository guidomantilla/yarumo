package breaker

import (
	"net/http"
)

// NoopFailOnResponse always returns false; disables response-based
// failure reporting. Only transport errors count against the breaker.
// This is the default value of WithFailOnResponse.
func NoopFailOnResponse(_ *http.Response) bool {
	return false
}

// FailOn5xxAnd429 returns true for 5xx server errors and 429 throttling.
// Other status codes are not reported as failures. Pass it to
// WithFailOnResponse to opt into HTTP-status-driven failure counting.
func FailOn5xxAnd429(res *http.Response) bool {
	if res == nil {
		return false
	}

	if res.StatusCode == http.StatusTooManyRequests {
		return true
	}

	return res.StatusCode >= 500 && res.StatusCode < 600
}
