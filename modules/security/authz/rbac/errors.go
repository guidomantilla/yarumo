package rbac

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// RBACType is the error domain identifier for RBAC failures.
const RBACType = "rbac"

var (
	_ error = (*Error)(nil)

	_ ErrRBACFn = ErrRBAC
)

// ErrRBACFn is the function type for ErrRBAC.
type ErrRBACFn func(causes ...error) error

// Sentinel errors for RBAC operations.
var (
	// ErrRBACFailed indicates that an RBAC operation failed.
	ErrRBACFailed = errors.New("rbac operation failed")
	// ErrRoleEmpty indicates that an empty role name was passed to a
	// configuration call (AddRole, Inherit, …).
	ErrRoleEmpty = errors.New("role name is empty")
	// ErrPermissionEmpty indicates that an empty permission string
	// was passed.
	ErrPermissionEmpty = errors.New("permission is empty")
	// ErrInheritanceCycle indicates that a role inheritance edge would
	// create a cycle (e.g. admin > editor > admin).
	ErrInheritanceCycle = errors.New("role inheritance cycle")
)

// Error is the domain error type for RBAC operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string including the type
// classification.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("rbac %s error: %s", e.Type, e.Err)
}

// ErrRBAC wraps the given causes into a domain Error for RBAC
// failures.
func ErrRBAC(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RBACType,
			Err:  errors.Join(append(causes, ErrRBACFailed)...),
		},
	}
}
