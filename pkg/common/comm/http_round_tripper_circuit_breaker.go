package comm

import (
	"net/http"

	gobreaker "github.com/sony/gobreaker/v2"
)

type HttpCircuitBreakerRoundTripper struct {
	Breaker *gobreaker.CircuitBreaker[*http.Response]
	Next    http.RoundTripper
}

func (tripper *HttpCircuitBreakerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := req.Clone(req.Context())
	return tripper.Breaker.Execute(func() (*http.Response, error) {
		return tripper.Next.RoundTrip(newReq)
	})
}
