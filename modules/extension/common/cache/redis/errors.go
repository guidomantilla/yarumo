package redis

import (
	"errors"
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
)

// CacheRedisType is the domain type tag attached to every Error produced for the redis backend.
const CacheRedisType = "cache-redis"

// Error is the domain error for redis cache backend operations.
type Error struct {
	cerrs.TypedError
}

// Error returns the formatted error string.
func (e *Error) Error() string {
	cassert.NotNil(e, "error is nil")
	cassert.NotNil(e.Err, "internal error is nil")

	return fmt.Sprintf("%s error: %s", e.Type, e.Err)
}

// Sentinel errors for redis backend failure modes.
var (
	ErrRedisCommandFailed = errors.New("redis command failed")
	ErrRedisEncodeFailed  = errors.New("redis codec encode failed")
	ErrRedisDecodeFailed  = errors.New("redis codec decode failed")
)

// ErrCommand creates a redis cache domain error joining the given causes with ErrRedisCommandFailed.
func ErrCommand(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CacheRedisType,
			Err:  errors.Join(append(causes, ErrRedisCommandFailed)...),
		},
	}
}

// ErrEncode creates a redis cache domain error joining the given causes with ErrRedisEncodeFailed.
func ErrEncode(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CacheRedisType,
			Err:  errors.Join(append(causes, ErrRedisEncodeFailed)...),
		},
	}
}

// ErrDecode creates a redis cache domain error joining the given causes with ErrRedisDecodeFailed.
func ErrDecode(causes ...error) error {
	return &Error{
		TypedError: cerrs.TypedError{
			Type: CacheRedisType,
			Err:  errors.Join(append(causes, ErrRedisDecodeFailed)...),
		},
	}
}
