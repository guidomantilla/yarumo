package authn

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// AuthnType is the error domain identifier for authentication failures.
const AuthnType = "authn"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for authentication operations. Concrete Authenticator
// implementations wrap one of these sentinels so transport middleware
// can translate failures to a uniform 401 response without inspecting
// impl-specific error types.
var (
	// ErrAuthenticationFailed is the umbrella sentinel that
	// ErrAuthentication always joins. errors.Is(err,
	// ErrAuthenticationFailed) returns true for every error produced by
	// this package.
	ErrAuthenticationFailed = errors.New("authentication failed")
	// ErrTokenEmpty indicates the caller provided an empty token
	// string. Returned by Authenticator.Validate before any
	// verification work is attempted.
	ErrTokenEmpty = errors.New("token is empty")
	// ErrTokenInvalid indicates the token failed verification
	// (signature, expiry, issuer mismatch, malformed claims, ...).
	// Concrete causes are joined into the same error via cerrs.Wrap.
	ErrTokenInvalid = errors.New("token is invalid")
	// ErrAuthenticatorNil indicates a nil Authenticator was supplied
	// to middleware or interceptor factories.
	ErrAuthenticatorNil = errors.New("authenticator is nil")
	// ErrHeaderMissing indicates the transport request carried no
	// authentication header at all.
	ErrHeaderMissing = errors.New("authentication header missing")
	// ErrHeaderMalformed indicates the authentication header was
	// present but did not match the expected "Bearer <token>" shape.
	ErrHeaderMalformed = errors.New("authentication header malformed")
	// ErrPrincipalNil indicates an Authenticator returned a nil
	// *Principal with no error. Treated by middleware as an invalid
	// token to keep the 401 contract uniform.
	ErrPrincipalNil = errors.New("principal is nil")
)

// Error is the domain error type for authentication operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("authn %s error: %s", e.Type, e.Err)
}

// ErrAuthentication wraps the given causes into a domain Error joined
// with ErrAuthenticationFailed. Concrete Authenticators and transport
// middleware funnel every failure through this factory so callers can
// match the whole family via errors.Is(err, ErrAuthenticationFailed).
func ErrAuthentication(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: AuthnType,
			Err:  errors.Join(append(causes, ErrAuthenticationFailed)...),
		},
	}
}
