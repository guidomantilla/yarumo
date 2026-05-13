// Package bayesian provides types shared across the Bayesian inference sub-packages.
package bayesian

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error domain type for Bayesian inference errors.
const (
	BayesianType = "inference-bayesian"
)

var _ error = (*Error)(nil)

// Error is the domain error for Bayesian inference operations.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for Bayesian inference failure modes.
var (
	ErrNetworkInvalid           = errors.New("network is invalid")
	ErrCyclicNetwork            = errors.New("network contains a cycle")
	ErrQueryNotInNetwork        = errors.New("query variable not in network")
	ErrBayesianQueryFailed      = errors.New("bayesian query failed")
	ErrBayesianValidationFailed = errors.New("bayesian validation failed")
)

// ErrQuery creates a Bayesian domain error joining the given causes.
func ErrQuery(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: BayesianType,
			Err:  errors.Join(append(errs, ErrBayesianQueryFailed)...),
		},
	}
}

// ErrValidation creates a Bayesian domain error joining the given causes.
func ErrValidation(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: BayesianType,
			Err:  errors.Join(append(errs, ErrBayesianValidationFailed)...),
		},
	}
}
