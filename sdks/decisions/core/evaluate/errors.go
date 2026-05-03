package evaluate

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// EvaluateType is the error type for evaluate engine errors.
const EvaluateType = "evaluate"

var _ error = (*Error)(nil)

// Error is the domain error type for the evaluate engine.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for operational failures.
var (
	ErrExecuteFailed = errors.New("execute failed")
	ErrExplainFailed = errors.New("explain failed")
	ErrAuditFailed   = errors.New("audit failed")
	ErrCascadeFailed = errors.New("cascade failed")
)

// Sentinel errors for dispatch-level failures.
var (
	ErrTypeMismatch  = errors.New("input type mismatch")
	ErrMissingConfig = errors.New("missing paradigm config")
	ErrUnsupported   = errors.New("unsupported paradigm")
	ErrNoBinder      = errors.New("no binder configured")
)

// Sentinel errors for model-level failures.
var (
	ErrInvalidHitPolicy = errors.New("invalid hit policy")
	ErrNoMatch          = errors.New("no matching rules")
	ErrMultipleMatches  = errors.New("multiple matches for unique policy")
	ErrConditionEval    = errors.New("condition evaluation failed")
)

// ErrExecute creates an execute error from the given causes.
func ErrExecute(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: EvaluateType,
			Err:  errors.Join(append(errs, ErrExecuteFailed)...),
		},
	}
}

// ErrExplain creates an explain error from the given causes.
func ErrExplain(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: EvaluateType,
			Err:  errors.Join(append(errs, ErrExplainFailed)...),
		},
	}
}

// ErrAudit creates an audit error from the given causes.
func ErrAudit(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: EvaluateType,
			Err:  errors.Join(append(errs, ErrAuditFailed)...),
		},
	}
}

// ErrCascade creates a cascade error from the given causes.
func ErrCascade(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: EvaluateType,
			Err:  errors.Join(append(errs, ErrCascadeFailed)...),
		},
	}
}
