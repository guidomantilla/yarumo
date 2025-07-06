package boot

import "github.com/guidomantilla/yarumo/pkg/common/pointer"

func Add(container *Container, key string, value any) {
	container.more[key] = value
}

func Get[T any](container *Container, key string) T {
	value, exists := container.more[key]
	if !exists {
		return pointer.Zero[T]()
	}

	typedValue, ok := value.(T)
	if !ok {
		return pointer.Zero[T]()
	}

	return typedValue
}
