package comm

import (
	"net/http"

	"github.com/guidomantilla/yarumo/pkg/common/utils"
)

type HttpHeadersRoundTripper struct {
	headers   http.Header
	overwrite bool
	next      http.RoundTripper
}

func (tripper *HttpHeadersRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := req.Clone(req.Context())
	for key, values := range tripper.headers {
		if tripper.overwrite || utils.Empty(newReq.Header.Get(key)) {
			for _, value := range values {
				newReq.Header.Set(key, value)
			}
		}
	}

	return tripper.next.RoundTrip(newReq)
}
