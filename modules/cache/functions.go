package cache

import (
	"errors"
	"fmt"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Sentinel errors for internal helpers (kept private to functions.go).
var (
	errOptionsNil = errors.New("cache options is nil")
)

// stringKey converts an arbitrary comparable key to the canonical string key
// used by every supported in-memory backend.
//
// It accepts string directly to avoid allocation and falls back to
// fmt.Sprintf("%v", ...) for any other comparable type. Every comparable Go
// value has a non-empty default rendering, so the conversion never fails — the
// helper returns only string to keep the closures in backends.go tight.
func stringKey[K comparable](key K) string {
	var anyKey any = key
	value, ok := anyKey.(string)
	if ok {
		return value
	}

	return fmt.Sprintf("%v", key)
}

// validateOptions ensures that the Options reference is usable.
func validateOptions(opts *Options) error {
	if opts == nil {
		return cerrs.Wrap(errOptionsNil)
	}
	return nil
}

// assertValue type-asserts raw into V, returning ErrSerialize when the
// underlying backend hands back a value whose concrete type does not match the
// cache's declared V parameter. Centralised so the factory closures share a
// single mismatch path and the unit test exercises the contract once.
func assertValue[V any](raw any) (V, error) {
	var zero V
	value, ok := raw.(V)
	if !ok {
		return zero, ErrSerialize(errors.New("value type mismatch"))
	}
	return value, nil
}
