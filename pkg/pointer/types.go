package pointer

var (
	_ ZeroFn[any]         = Zero[any]
	_ IsZeroFn[any]       = IsZero[any]
	_ IsNotZeroFn[any]    = IsNotZero[any]
	_ NilFn[any]          = Nil[any]
	_ IsNilFn             = IsNil
	_ IsNotNilFn          = IsNotNil
	_ ToPtrFn[any]        = ToPtr[any]
	_ FromPtrFn[any]      = FromPtr[any]
	_ ToSlicePtrFn[any]   = ToSlicePtr[any]
	_ FromSlicePtrFn[any] = FromSlicePtr[any]
)

type ZeroFn[T any] func() T

type IsZeroFn[T comparable] func(v T) bool

type IsNotZeroFn[T comparable] func(v T) bool

type NilFn[T any] func() *T

type IsNilFn func(x any) bool

type IsNotNilFn func(x any) bool

type ToPtrFn[T any] func(x T) *T

type FromPtrFn[T any] func(x *T) T

type ToSlicePtrFn[T any] func(collection []T) []*T

type FromSlicePtrFn[T any] func(collection []*T) []T
