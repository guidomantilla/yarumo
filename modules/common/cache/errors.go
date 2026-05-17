package cache

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// CacheType is the domain type tag attached to every Error produced by this package.
const CacheType = "cache"

// Error is the domain error for cache operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("%s error: %s", e.Type, e.Err)
}

// Sentinel errors for cache failure modes.
var (
	ErrCacheTypeAssertion = errors.New("cache type assertion failed")
	ErrCacheMiss          = errors.New("cache miss")
	ErrCacheFailed        = errors.New("cache operation failed")
	ErrCacheNotRegistered = errors.New("cache not registered")
)

// ErrCache creates a cache domain error joining the given causes with ErrCacheFailed.
func ErrCache(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CacheType,
			Err:  errors.Join(append(causes, ErrCacheFailed)...),
		},
	}
}

// ErrTypeAssertion creates a cache domain error joining the given causes with ErrCacheTypeAssertion.
func ErrTypeAssertion(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CacheType,
			Err:  errors.Join(append(causes, ErrCacheTypeAssertion)...),
		},
	}
}

// ErrMiss creates a cache domain error joining the given causes with ErrCacheMiss.
func ErrMiss(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CacheType,
			Err:  errors.Join(append(causes, ErrCacheMiss)...),
		},
	}
}

// ErrNotRegistered creates a cache domain error indicating that no cache is
// registered under the given name.
func ErrNotRegistered(name string) error {
	cassert.NotEmpty(name, "name is empty")

	return &Error{
		TypedError: cerrs.TypedError{
			Type: CacheType,
			Err:  errors.Join(fmt.Errorf("cache %q", name), ErrCacheNotRegistered),
		},
	}
}
