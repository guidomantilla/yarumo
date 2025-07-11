package maths

import "github.com/guidomantilla/yarumo/pkg/common/maths/logic/predicates"

type List[T comparable] []T

func NewList[T comparable](elements ...T) List[T] {
	return elements
}

func (l List[T]) Concat(other List[T]) List[T] {
	return append(l, other...)
}

func (l List[T]) Length() int {
	return len(l)
}

func (l List[T]) Reverse() List[T] {
	n := len(l)
	out := make(List[T], n)
	for i := range l {
		out[n-1-i] = l[i]
	}
	return out
}

func (l List[T]) Size() int {
	return len(l)
}

func (l List[T]) Filter(pred predicates.Predicate[T]) List[T] {
	out := make(List[T], 0)
	for _, v := range l {
		if pred(v) {
			out = append(out, v)
		}
	}
	return out
}

func (l List[T]) Equal(other List[T]) bool {
	if len(l) != len(other) {
		return false
	}
	for i := range l {
		if l[i] != other[i] {
			return false
		}
	}
	return true
}

func (l List[T]) Empty() bool {
	return len(l) == 0
}

func (l List[T]) ToSlice() []T {
	out := make([]T, 0, len(l))
	for _, v := range l {
		out = append(out, v)
	}
	return out
}
