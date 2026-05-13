package cache

import (
	"errors"
	"fmt"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Sentinel errors for internal helpers (kept private to functions.go).
var (
	errOptionsNil    = errors.New("cache options is nil")
	errKeyConversion = errors.New("cache key conversion failed")
)

// stringKey converts an arbitrary comparable key to the canonical string key
// used by every supported gocache backend.
//
// It accepts string directly to avoid allocation and falls back to
// fmt.Sprintf("%v", ...) for any other comparable type.
func stringKey[K comparable](key K) (string, error) {
	var anyKey any = key
	value, ok := anyKey.(string)
	if ok {
		return value, nil
	}

	rendered := fmt.Sprintf("%v", key)
	if rendered == "" {
		return "", cerrs.Wrap(errKeyConversion, fmt.Errorf("empty key for %T", key))
	}

	return rendered, nil
}

// validateOptions ensures that the Options reference is usable.
func validateOptions(opts *Options) error {
	if opts == nil {
		return cerrs.Wrap(errOptionsNil)
	}
	return nil
}
