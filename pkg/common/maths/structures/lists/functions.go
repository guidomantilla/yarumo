package lists

import (
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/predicates"
)

type List[T any] []T

// Concat returns the concatenation of two slices: A ++ B
func Concat[T any](a, b List[T]) List[T] {
	return append(a, b...)
}

// Length returns the length of a slice
func Length[T any](s List[T]) int {
	return len(s)
}

// Reverse returns a new slice with elements in reverse order
func Reverse[T any](s List[T]) List[T] {
	n := len(s)
	out := make(List[T], n)
	for i := range s {
		out[n-1-i] = s[i]
	}
	return out
}

// Map applies a function to each element of the slice
func Map[T any, R any](s List[T], f func(T) R) []R {
	out := make([]R, len(s))
	for i, v := range s {
		out[i] = f(v)
	}
	return out
}

// Filter returns a new slice containing elements for which pred(elem) is true
func Filter[T any](s List[T], pred predicates.Predicate[T]) List[T] {
	out := make(List[T], 0)
	for _, v := range s {
		if pred(v) {
			out = append(out, v)
		}
	}
	return out
}

// Fold applies a reducer function from left to right
func Fold[T any, R any](s List[T], init R, f func(R, T) R) R {
	acc := init
	for _, v := range s {
		acc = f(acc, v)
	}
	return acc
}

// Equal compares two slices for element-wise equality (including order)
func Equal[T comparable](a, b List[T]) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
