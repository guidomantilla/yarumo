package comm

import (
	"net/http"

	gobreaker "github.com/sony/gobreaker"
)

type HttpCircuitBreakerRoundTripper struct {
	Breaker *gobreaker.CircuitBreaker
	Next    http.RoundTripper
}

func (tripper *HttpCircuitBreakerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := req.Clone(req.Context())
	res, err := tripper.Breaker.Execute(func() (any, error) {
		return tripper.Next.RoundTrip(newReq)
	})
	if err != nil {
		return nil, err
	}

	return res.(*http.Response), nil
}
