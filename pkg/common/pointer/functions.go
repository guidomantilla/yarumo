package pointer

import (
	"reflect"

	"github.com/guidomantilla/yarumo/pkg/common/constraints"
)

func IsPointer(v any) bool {
	if IsNil(v) {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Ptr
}

func IsType(v any, typeName string) bool {
	if IsNil(v) {
		return false
	}
	if IsPointer(v) {
		v = reflect.ValueOf(v).Elem().Interface()
		return reflect.TypeOf(v).String() == typeName
	}
	return reflect.TypeOf(v).String() == typeName
}

// Zero returns the zero value
func Zero[T any]() T {
	var zero T
	return zero
}

// IsZero returns true if argument is a zero value.
func IsZero[T constraints.Comparable](v T) bool {
	return Zero[T]() == v
}

// IsNotZero returns true if argument is not a zero value.
func IsNotZero[T constraints.Comparable](v T) bool {
	return Zero[T]() != v
}

// Nil returns a nil pointer of type.
func Nil[T any]() *T {
	return nil
}

// IsNil checks if a value is nil or if it's a reference type with a nil underlying value.
func IsNil(x any) bool {
	defer func() { recover() }() // nolint:errcheck
	return x == nil || reflect.ValueOf(x).IsNil()
}

// IsNotNil checks if a value is not nil or if it's not a reference type with a nil underlying value.
func IsNotNil(x any) bool {
	return !IsNil(x)
}

// ToPtr returns a pointer copy of value.
func ToPtr[T any](x T) *T {
	return &x
}

// FromPtr returns the pointer value or empty.
func FromPtr[T any](x *T) T {
	if x == nil {
		return Zero[T]()
	}

	return *x
}

// ToSlicePtr returns a slice of pointer copy of value.
func ToSlicePtr[T any](collection []T) []*T {
	result := make([]*T, len(collection))

	for i := range collection {
		result[i] = &collection[i]
	}
	return result
}

// FromSlicePtr returns a slice with the pointer values.
// Returns a zero value in case of a nil pointer element.
func FromSlicePtr[T any](collection []*T) []T {
	return convert(collection, func(x *T, _ int) T {
		if x == nil {
			return Zero[T]()
		}
		return *x
	})
}

// convert manipulates a slice and transforms it to a slice of another type.
func convert[T any, R any](collection []T, iteratee func(item T, index int) R) []R {
	result := make([]R, len(collection))

	for i := range collection {
		result[i] = iteratee(collection[i], i)
	}

	return result
}
