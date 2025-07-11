package maths

type Bag[T comparable] map[T]int

func New[T comparable](items ...T) Bag[T] {
	b := make(Bag[T])
	for _, item := range items {
		b[item]++
	}
	return b
}

func (b Bag[T]) Add(item T, count int) {
	if count <= 0 {
		return
	}
	b[item] += count
}

func (b Bag[T]) Remove(item T, count int) {
	if current, ok := b[item]; ok {
		if count >= current {
			delete(b, item)
		} else {
			b[item] -= count
		}
	}
}

func (b Bag[T]) Count(item T) int {
	return b[item]
}

func (b Bag[T]) UniqueSize() int {
	return len(b)
}

func (b Bag[T]) Union(other Bag[T]) Bag[T] {
	out := make(Bag[T])
	for item, count := range b {
		out[item] = count
	}
	for item, count := range other {
		out[item] += count
	}
	return out
}

func (b Bag[T]) Intersection(other Bag[T]) Bag[T] {
	out := make(Bag[T])
	for item, count := range b {
		if otherCount, ok := other[item]; ok {
			out[item] = min(count, otherCount)
		}
	}
	return out
}

func (b Bag[T]) Difference(other Bag[T]) Bag[T] {
	out := make(Bag[T])
	for item, count := range b {
		if otherCount, ok := other[item]; !ok {
			out[item] = count
		} else if count > otherCount {
			out[item] = count - otherCount
		}
	}
	return out
}

func (b Bag[T]) Complement(universe Bag[T]) Bag[T] {
	out := make(Bag[T])
	for item, count := range universe {
		if currentCount, ok := b[item]; !ok || currentCount < count {
			out[item] = count
		}
	}
	return out
}

func (b Bag[T]) IsSubset(other Bag[T]) bool {
	for item, count := range other {
		if otherCount, ok := b[item]; !ok || otherCount > count {
			return false
		}
	}
	return true
}

func (b Bag[T]) Size() int {
	size := 0
	for _, count := range b {
		size += count
	}
	return size
}

func (b Bag[T]) Filter(pred Predicate[T]) Bag[T] {
	out := make(Bag[T])
	for item, count := range b {
		if pred(item) {
			out[item] = count
		}
	}
	return out
}

func (b Bag[T]) Equal(other Bag[T]) bool {
	if len(b) != len(other) {
		return false
	}
	for item, count := range b {
		if otherCount, ok := other[item]; !ok || count != otherCount {
			return false
		}
	}
	return true
}

func (b Bag[T]) Empty() bool {
	return len(b) == 0
}

func (b Bag[T]) ToSlice() []T {
	out := make([]T, 0, b.Size())
	for item, count := range b {
		for i := 0; i < count; i++ {
			out = append(out, item)
		}
	}
	return out
}
