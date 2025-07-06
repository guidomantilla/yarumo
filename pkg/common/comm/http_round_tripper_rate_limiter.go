package comm

import (
	"net/http"

	"golang.org/x/time/rate"
)

type HttpRateLimiterRoundTripper struct {
	RateLimiter *rate.Limiter
	Next        http.RoundTripper
}

func (tripper *HttpRateLimiterRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := req.Clone(req.Context())
	err := tripper.RateLimiter.Wait(newReq.Context())
	if err != nil {
		return nil, err
	}
	return tripper.Next.RoundTrip(newReq)
}
