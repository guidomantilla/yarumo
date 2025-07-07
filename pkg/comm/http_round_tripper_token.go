package comm

import (
	"fmt"
	"net/http"

	"github.com/guidomantilla/yarumo/pkg/common/utils"
)

type HttpTokenRoundTripper struct {
	token      string     //nolint:unused
	headerName string     // e.g. "Authorization"
	scheme     string     // e.g. "Bearer", or "" if raw token
	getToken   GetTokenFn // Function to get the token dynamically
	next       http.RoundTripper
}

func (tripper *HttpTokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := req.Clone(req.Context())
	token, err := tripper.getToken(newReq)
	if err != nil {
		return nil, err
	}

	newReq.Header.Set(tripper.headerName, *token)
	if utils.NotEmpty(tripper.scheme) {
		newReq.Header.Set(tripper.headerName, fmt.Sprintf("%s %s", tripper.scheme, *token))
	}

	return tripper.next.RoundTrip(newReq)
}
