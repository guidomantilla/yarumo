package authz

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// AuthzType is the error domain identifier for authorization failures.
const AuthzType = "authz"

var (
	_ error = (*Error)(nil)
)

// Sentinel errors for authorization failures.
var (
	// ErrAuthzFailed indicates that authorization failed.
	ErrAuthzFailed = errors.New("authorization failed")
	// ErrDenied indicates that a Policy returned Deny for the Request.
	ErrDenied = errors.New("access denied")
	// ErrAbstained indicates that a Policy returned Abstain and no
	// fallback policy granted access; treated as deny by Require.
	ErrAbstained = errors.New("policy abstained")
	// ErrPolicyNil indicates that a nil Policy was passed to Require.
	ErrPolicyNil = errors.New("policy is nil")
	// ErrPrincipalNil indicates that PrincipalReader returned no
	// principal for the request.
	ErrPrincipalNil = errors.New("principal is nil")
	// ErrPrincipalReaderNil indicates that Require was configured with
	// a nil PrincipalReader.
	ErrPrincipalReaderNil = errors.New("principal reader is nil")
	// ErrActionEmpty indicates that Require was configured with an
	// empty action string.
	ErrActionEmpty = errors.New("action is empty")
)

// Error is the domain error type for authorization failures.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("authz %s error: %s", e.Type, e.Err)
}

// ErrAuthz wraps the given causes into a domain Error for authorization
// failures.
func ErrAuthz(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: AuthzType,
			Err:  errors.Join(append(causes, ErrAuthzFailed)...),
		},
	}
}
