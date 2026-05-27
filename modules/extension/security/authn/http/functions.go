package http

import (
	"net/http"
	"strings"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/core/security/authn"
)

// NewMiddleware returns a server-side net/http Middleware that
// terminates Bearer authentication against the given Authenticator.
//
// On success the *Principal is injected into the request ctx via
// authn.WithPrincipal and the wrapped handler is invoked. On failure
// the configured ErrorHandler shapes the response; the wrapped handler
// is never reached. The default ErrorHandler writes 401 Unauthorized
// with an empty body.
//
// A nil authenticator panics via cassert.NotNil so construction-time
// wiring mistakes surface immediately rather than as silent 401s at
// runtime.
func NewMiddleware(authenticator authn.Authenticator, options ...Option) Middleware {
	cassert.NotNil(authenticator, "authenticator is nil")

	opts := NewOptions(options...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := extractBearerToken(r, opts.headerName, opts.scheme)
			if err != nil {
				opts.errorHandler(w, r, authn.ErrAuthentication(err))

				return
			}

			principal, err := authenticator.Validate(r.Context(), token)
			if err != nil {
				opts.errorHandler(w, r, err)

				return
			}

			if principal == nil {
				opts.errorHandler(w, r, authn.ErrAuthentication(authn.ErrPrincipalNil))

				return
			}

			ctx := authn.WithPrincipal(r.Context(), principal)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractBearerToken returns the credential portion of an
// "Authorization: Bearer <token>" header. Comparisons against the
// scheme are case-insensitive per RFC 7235; the token portion is
// returned verbatim. Missing headers return ErrHeaderMissing; any other
// shape (no whitespace separator, wrong scheme, empty credential)
// returns ErrHeaderMalformed.
func extractBearerToken(r *http.Request, headerName, scheme string) (string, error) {
	value := r.Header.Get(headerName)
	if value == "" {
		return "", authn.ErrHeaderMissing
	}

	scheme = strings.ToLower(scheme)

	parts := strings.SplitN(value, " ", 2)
	if len(parts) != 2 {
		return "", authn.ErrHeaderMalformed
	}

	if !strings.EqualFold(parts[0], scheme) {
		return "", authn.ErrHeaderMalformed
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", authn.ErrHeaderMalformed
	}

	return token, nil
}
