// Package pointer provides utilities for pointer manipulation, nil checking, and type introspection.
package pointer

import (
	cconstraints "github.com/guidomantilla/yarumo/common/constraints"
)

var (
	_ IsEmptyFn           = IsEmpty
	_ IsNotEmptyFn        = IsNotEmpty
	_ IsPointerFn         = IsPointer
	_ IsNotPointerFn      = IsNotPointer
	_ IsTypeFn            = IsType
	_ IsStructFn          = IsStruct
	_ IsChanFn            = IsChan
	_ IsSliceFn           = IsSlice
	_ IsMapFn             = IsMap
	_ ZeroFn[any]         = Zero
	_ IsZeroFn[any]       = IsZero
	_ IsNotZeroFn[any]    = IsNotZero
	_ IsNilFn             = IsNil
	_ IsNotNilFn          = IsNotNil
	_ ToPtrFn[any]        = ToPtr
	_ FromPtrFn[any]      = FromPtr
	_ ToSlicePtrFn[any]   = ToSlicePtr
	_ FromSlicePtrFn[any] = FromSlicePtr
)

// IsEmptyFn is the function type for IsEmpty.
type IsEmptyFn func(x any) bool

// IsNotEmptyFn is the function type for IsNotEmpty.
type IsNotEmptyFn func(x any) bool

// IsPointerFn is the function type for IsPointer.
type IsPointerFn func(x any) bool

// IsNotPointerFn is the function type for IsNotPointer.
type IsNotPointerFn func(x any) bool

// IsTypeFn is the function type for IsType.
type IsTypeFn func(v any, typeName string) bool

// IsStructFn is the function type for IsStruct.
type IsStructFn func(x any) bool

// IsChanFn is the function type for IsChan.
type IsChanFn func(x any) bool

// IsSliceFn is the function type for IsSlice.
type IsSliceFn func(x any) bool

// IsMapFn is the function type for IsMap.
type IsMapFn func(x any) bool

// ZeroFn is the function type for Zero.
type ZeroFn[T any] func() T

// IsZeroFn is the function type for IsZero.
type IsZeroFn[T cconstraints.Comparable] func(v T) bool

// IsNotZeroFn is the function type for IsNotZero.
type IsNotZeroFn[T cconstraints.Comparable] func(v T) bool

// IsNilFn is the function type for IsNil.
type IsNilFn func(x any) bool

// IsNotNilFn is the function type for IsNotNil.
type IsNotNilFn func(x any) bool

// ToPtrFn is the function type for ToPtr.
type ToPtrFn[T any] func(x T) *T

// FromPtrFn is the function type for FromPtr.
type FromPtrFn[T any] func(x *T) T

// ToSlicePtrFn is the function type for ToSlicePtr.
type ToSlicePtrFn[T any] func(collection []T) []*T

// FromSlicePtrFn is the function type for FromSlicePtr.
type FromSlicePtrFn[T any] func(collection []*T) []T
