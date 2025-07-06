package comm

import (
	"net/http"

	"github.com/guidomantilla/yarumo/pkg/common/utils"
)

type HttpHeadersRoundTripper struct {
	Headers   http.Header
	Overwrite bool
	Next      http.RoundTripper
}

func (tripper *HttpHeadersRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := req.Clone(req.Context())
	for key, values := range tripper.Headers {
		if tripper.Overwrite || utils.Empty(newReq.Header.Get(key)) {
			for _, value := range values {
				newReq.Header.Set(key, value)
			}
		}
	}

	return tripper.Next.RoundTrip(newReq)
}
