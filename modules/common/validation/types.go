// Package validation provides function-typed leaf validators and reflection
// helpers used by callers (handlers, services) and by higher-level
// config-driven engines such as modules/validation/.
//
// Leaves are plain functions: they take a value (and any extra parameters)
// and return an error. They never depend on struct tags. The package error
// type embeds errs.TypedError with Type = "validation" so violations group
// cleanly under errs.AsErrorInfo.
package validation

import (
	cconstraints "github.com/guidomantilla/yarumo/common/constraints"
)

// Numeric is the constraint accepted by the numeric leaf validators. It
// covers every Go integer and floating-point type.
type Numeric interface {
	cconstraints.Number
}

// CheckFn is the function type for a generic leaf validator over a single value.
type CheckFn[T any] func(value T) error

// CollectionCheckFn is the function type for a leaf validator over a slice.
type CollectionCheckFn[T any] func(xs []T) error

// FieldFn is the function type for GetField.
type FieldFn func(obj any, path string) (any, error)

var (
	_ FieldFn = GetField
)
