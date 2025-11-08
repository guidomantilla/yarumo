package pointer

import (
	"github.com/guidomantilla/yarumo/modules/common/constraints"
)

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

type IsZeroFn[T constraints.Comparable] func(v T) bool

type IsNotZeroFn[T constraints.Comparable] func(v T) bool

type NilFn[T any] func() *T

type IsNilFn func(x any) bool

type IsNotNilFn func(x any) bool

type ToPtrFn[T any] func(x T) *T

type FromPtrFn[T any] func(x *T) T

type ToSlicePtrFn[T any] func(collection []T) []*T

type FromSlicePtrFn[T any] func(collection []*T) []T
