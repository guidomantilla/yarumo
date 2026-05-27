package rules

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// Error domain type for rule errors.
const (
	RuleType = "inference-rule"
)

var _ error = (*Error)(nil)

// Error is the domain error for rule operations.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for rule failure modes.
var (
	ErrRuleInvalid          = errors.New("rule is invalid")
	ErrRuleValidationFailed = errors.New("rule validation failed")
)

// ErrValidation creates a rule domain error joining the given causes.
func ErrValidation(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RuleType,
			Err:  errors.Join(append(errs, ErrRuleValidationFailed)...),
		},
	}
}
