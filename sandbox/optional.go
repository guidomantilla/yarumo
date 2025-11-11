package sandbox

type Optional[T any] struct {
	ok    bool
	value T
}

func NewOptional[T any](value T) *Optional[T] {
	return &Optional[T]{
		ok:    true,
		value: value,
	}
}

func (o *Optional[T]) HasValue() bool {
	return o.ok
}

func (o *Optional[T]) Value() T {
	return o.value
}

func (o *Optional[T]) Set(value T) {
	o.ok = true
	o.value = value
}

func (o *Optional[T]) Get() (T, bool) {
	return o.value, o.ok
}

func (o *Optional[T]) Clear() {
	o.ok = false
}

func (o *Optional[T]) Default(value T) T {
	if o.ok {
		return o.value
	}
	return value
}
