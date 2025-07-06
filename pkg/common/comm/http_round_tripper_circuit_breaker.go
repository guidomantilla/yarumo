package comm

import (
	"net/http"

	"github.com/sony/gobreaker"
)

type HttpCircuitBreakerRoundTripper struct {
	Breaker *gobreaker.CircuitBreaker
	Next    http.RoundTripper
}

func (tripper *HttpCircuitBreakerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := req.Clone(req.Context())
	var resp *http.Response
	_, err := tripper.Breaker.Execute(func() (any, error) {
		r, err := tripper.Next.RoundTrip(newReq)
		resp = r
		return r, err
	})
	return resp, err
}
