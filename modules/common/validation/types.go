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

var (
	_ CheckFn[string]           = IsEmail
	_ CheckFn[string]           = IsURL
	_ CheckFn[string]           = IsUUID
	_ CheckFn[string]           = IsULID
	_ CheckFn[string]           = IsRequired[string]
	_ CheckFn[string]           = MustBeUndefined[string]
	_ MinLenFn                  = MinLen
	_ MaxLenFn                  = MaxLen
	_ MatchesRegexFn            = MatchesRegex
	_ MinFn[int]                = Min[int]
	_ MaxFn[int]                = Max[int]
	_ InRangeFn[int]            = InRange[int]
	_ CollectionCheckFn[string] = NonEmpty[string]
	_ EachFn[string]            = Each[string]
	_ FieldFn                   = GetField
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

// MinLenFn is the function type for MinLen.
type MinLenFn func(s string, n int) error

// MaxLenFn is the function type for MaxLen.
type MaxLenFn func(s string, n int) error

// MatchesRegexFn is the function type for MatchesRegex.
type MatchesRegexFn func(s string, pattern string) error

// MinFn is the function type for Min.
type MinFn[T Numeric] func(v, lo T) error

// MaxFn is the function type for Max.
type MaxFn[T Numeric] func(v, hi T) error

// InRangeFn is the function type for InRange.
type InRangeFn[T Numeric] func(v, lo, hi T) error

// EachFn is the function type for Each.
type EachFn[T any] func(xs []T, check func(T) error) error

// FieldFn is the function type for GetField.
type FieldFn func(obj any, path string) (any, error)
