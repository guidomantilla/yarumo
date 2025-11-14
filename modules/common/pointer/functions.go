package pointer

import (
	"reflect"

	"github.com/guidomantilla/yarumo/common/constraints"
)

// IsNil checks if a value is nil or if it's a reference type with a nil underlying value.
func IsNil(x any) bool {
	if x == nil {
		return true
	}
	v := reflect.ValueOf(x)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Slice, reflect.Array, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}

// IsNotNil checks if a value is not nil or if it's not a reference type with a nil underlying value.
func IsNotNil(x any) bool {
	return !IsNil(x)
}

// IsEmpty checks if a value is empty.
func IsEmpty(x any) bool {
	if x == nil {
		return true
	}

	val := reflect.ValueOf(x)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return true
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return val.Len() == 0
	default:
		return val.IsZero()
	}
}

// IsNotEmpty checks if a value is not empty.
func IsNotEmpty(x any) bool {
	return !IsEmpty(x)
}

// IsPointer checks if the value is a pointer type.
func IsPointer(v any) bool {
	if IsNil(v) {
		return false
	}
	return reflect.ValueOf(v).Kind() == reflect.Ptr
}

// IsNotPointer checks if the value is not a pointer type.
func IsNotPointer(v any) bool {
	return !IsPointer(v)
}

// Zero returns the zero value
func Zero[T any]() T {
	var zero T
	return zero
}

// IsZero returns true if the argument is a zero value.
func IsZero[T constraints.Comparable](v T) bool {
	return Zero[T]() == v
}

// IsNotZero returns true if the argument is not a zero value.
func IsNotZero[T constraints.Comparable](v T) bool {
	return Zero[T]() != v
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
// Returns a zero value in the case of a nil pointer element.
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

// IsType checks if the value is of a specific type.
func IsType(v any, typeName string) bool {
	if v == nil {
		return false
	}
	if IsPointer(v) {
		v = reflect.ValueOf(v).Elem().Interface()
		return reflect.TypeOf(v).String() == typeName
	}
	return reflect.TypeOf(v).String() == typeName
}

// IsStruct checks if the value is a struct
func IsStruct(x any) bool {
	if x == nil {
		return false
	}

	val := reflect.ValueOf(x)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	return val.Kind() == reflect.Struct
}

// IsChan checks if the value is a channel
func IsChan(x any) bool {
	if x == nil {
		return false
	}

	val := reflect.ValueOf(x)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	return val.Kind() == reflect.Chan
}

// IsSlice checks if the value is a slice or an array
func IsSlice(x any) bool {
	if x == nil {
		return false
	}

	val := reflect.ValueOf(x)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	return val.Kind() == reflect.Slice || val.Kind() == reflect.Array
}

// IsMap checks if the value is a map
func IsMap(x any) bool {
	if x == nil {
		return false
	}
	val := reflect.ValueOf(x)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	return val.Kind() == reflect.Map
}
