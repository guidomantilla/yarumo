package jwt

import (
	"errors"
)

// Sentinel errors specific to the JWT Authenticator. They are joined
// via authn.ErrAuthentication(...) so callers can match the umbrella
// authn.ErrAuthenticationFailed AND a precise reason.
var (
	// ErrMethodNil indicates a nil *tokens.Method was supplied to
	// NewJWTAuthenticator.
	ErrMethodNil = errors.New("tokens method is nil")
	// ErrSubjectClaimMissing indicates the resolved subject-claim key
	// was not present in the JWT payload, or its value was not a
	// non-empty string. The Principal cannot be constructed without an
	// ID.
	ErrSubjectClaimMissing = errors.New("subject claim missing in token payload")
)
