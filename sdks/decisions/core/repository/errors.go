package repository

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// RepositoryType is the error type for repository errors.
const RepositoryType = "repository"

var _ error = (*Error)(nil)

// Error is the domain error type for the repository package.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for repository operations.
var (
	ErrNotFound   = errors.New("ruleset not found")
	ErrGetFailed  = errors.New("get failed")
	ErrListFailed = errors.New("list failed")
	ErrSaveFailed = errors.New("save failed")
)

// ErrGet creates a get error from the given causes.
func ErrGet(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RepositoryType,
			Err:  errors.Join(append(errs, ErrGetFailed)...),
		},
	}
}

// ErrList creates a list error from the given causes.
func ErrList(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RepositoryType,
			Err:  errors.Join(append(errs, ErrListFailed)...),
		},
	}
}

// ErrSave creates a save error from the given causes.
func ErrSave(errs ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: RepositoryType,
			Err:  errors.Join(append(errs, ErrSaveFailed)...),
		},
	}
}
