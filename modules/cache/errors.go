package cache

import (
	"errors"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// Error domain type for cache operation errors.
const (
	CacheType = "cache"
)

var _ error = (*Error)(nil)

// Error is the domain error for cache operations.
type Error struct {
	cerrs.TypedError
}

// Sentinel errors for cache failure modes.
var (
	ErrCacheMiss          = errors.New("cache miss")
	ErrSerialization      = errors.New("serialization failed")
	ErrBackendUnavailable = errors.New("cache backend unavailable")
	ErrCacheFailed        = errors.New("cache operation failed")
	ErrUnsupportedBackend = errors.New("unsupported cache backend")
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

// ErrMiss creates a cache domain error joining the given causes with ErrCacheMiss.
func ErrMiss(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CacheType,
			Err:  errors.Join(append(causes, ErrCacheMiss)...),
		},
	}
}

// ErrSerialize creates a cache domain error joining the given causes with ErrSerialization.
func ErrSerialize(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CacheType,
			Err:  errors.Join(append(causes, ErrSerialization)...),
		},
	}
}

// ErrBackend creates a cache domain error joining the given causes with ErrBackendUnavailable.
func ErrBackend(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CacheType,
			Err:  errors.Join(append(causes, ErrBackendUnavailable)...),
		},
	}
}

// ErrUnsupported creates a cache domain error joining the given causes with ErrUnsupportedBackend.
func ErrUnsupported(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CacheType,
			Err:  errors.Join(append(causes, ErrUnsupportedBackend)...),
		},
	}
}
