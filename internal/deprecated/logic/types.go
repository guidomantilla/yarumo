package logic

import (
	"cmp"
)

// Signed is a constraint that permits any signed integer type.
// If future releases of Go add new predeclared signed integer types,
// this constraint will be modified to include them.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Unsigned is a constraint that permits any unsigned integer type.
// If future releases of Go add new predeclared unsigned integer types,
// this constraint will be modified to include them.
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Integer is a constraint that permits any integer type.
// If future releases of Go add new predeclared integer types,
// this constraint will be modified to include them.
type Integer interface {
	Signed | Unsigned
}

// Float is a constraint that permits any floating-point type.
// If future releases of Go add new predeclared floating-point types,
// this constraint will be modified to include them.
type Float interface {
	~float32 | ~float64
}

// Complex is a constraint that permits any complex numeric type.
// If future releases of Go add new predeclared complex numeric types,
// this constraint will be modified to include them.
type Complex interface {
	~complex64 | ~complex128
}

/**/

type Comparable = comparable

type Ordenable = cmp.Ordered

type Number interface {
	Integer | Float
}

var (
	_ IsPointerFn         = IsPointer
	_ IsTypeFn            = IsType
	_ ZeroFn[any]         = Zero
	_ IsZeroFn[any]       = IsZero
	_ IsNotZeroFn[any]    = IsNotZero
	_ NilFn[any]          = Nil
	_ IsNilFn             = IsNil
	_ IsNotNilFn          = IsNotNil
	_ ToPtrFn[any]        = ToPtr
	_ FromPtrFn[any]      = FromPtr
	_ ToSlicePtrFn[any]   = ToSlicePtr
	_ FromSlicePtrFn[any] = FromSlicePtr
)

type IsPointerFn func(x any) bool

type IsTypeFn func(v any, typeName string) bool

type ZeroFn[T any] func() T

type IsZeroFn[T Comparable] func(v T) bool

type IsNotZeroFn[T Comparable] func(v T) bool

type NilFn[T any] func() *T

type IsNilFn func(x any) bool

type IsNotNilFn func(x any) bool

type ToPtrFn[T any] func(x T) *T

type FromPtrFn[T any] func(x *T) T

type ToSlicePtrFn[T any] func(collection []T) []*T

type FromSlicePtrFn[T any] func(collection []*T) []T
