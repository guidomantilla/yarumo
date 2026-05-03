package adapters

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// AdapterType is the error type for adapter errors.
const AdapterType = "adapter"

var _ error = (*Error)(nil)

// Error is the domain error type for the adapters package.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for adapter operations.
var (
	ErrAdaptRulesFailed      = errors.New("adapt deductive rules failed")
	ErrAdaptNetworkFailed    = errors.New("adapt bayesian network failed")
	ErrAdaptVariablesFailed  = errors.New("adapt fuzzy variables failed")
	ErrAdaptMembershipFailed = errors.New("adapt membership function failed")
	ErrInvalidParamCount     = errors.New("invalid parameter count")
	ErrUnknownMembershipType = errors.New("unknown membership function type")
)

// ErrAdaptRules creates an adapt-rules error from the given causes.
func ErrAdaptRules(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: AdapterType,
			Err:  errors.Join(append(errs, ErrAdaptRulesFailed)...),
		},
	}
}

// ErrAdaptNetwork creates an adapt-network error from the given causes.
func ErrAdaptNetwork(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: AdapterType,
			Err:  errors.Join(append(errs, ErrAdaptNetworkFailed)...),
		},
	}
}

// ErrAdaptVariables creates an adapt-variables error from the given causes.
func ErrAdaptVariables(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: AdapterType,
			Err:  errors.Join(append(errs, ErrAdaptVariablesFailed)...),
		},
	}
}

// ErrAdaptMembership creates an adapt-membership error from the given causes.
func ErrAdaptMembership(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: AdapterType,
			Err:  errors.Join(append(errs, ErrAdaptMembershipFailed)...),
		},
	}
}
