package comm

import (
	"fmt"
	"net/http"

	"github.com/guidomantilla/yarumo/pkg/common/utils"
)

type HttpTokenRoundTripper struct {
	Token      string
	HeaderName string     // e.g. "Authorization"
	Scheme     string     // e.g. "Bearer", or "" if raw token
	GetToken   GetTokenFn // Function to get the token dynamically
	Next       http.RoundTripper
}

func (tripper *HttpTokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := req.Clone(req.Context())
	token, err := tripper.GetToken(newReq)
	if err != nil {
		return nil, err
	}

	newReq.Header.Set(tripper.HeaderName, *token)
	if utils.NotEmpty(tripper.Scheme) {
		newReq.Header.Set(tripper.HeaderName, fmt.Sprintf("%s %s", tripper.Scheme, *token))
	}

	return tripper.Next.RoundTrip(newReq)
}
