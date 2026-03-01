// Package hmacs provides HMAC (Hash-based Message Authentication Code) key
// generation, digest computation, and validation using pluggable hash functions
// and a thread-safe name-based registry.
package hmacs

import (
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// Type compliance.
var (
	_ KeyFn      = key
	_ DigestFn   = digest
	_ ValidateFn = validate
)

// KeyFn is the function type for key generation.
type KeyFn func(method *Method) (ctypes.Bytes, error)

// DigestFn is the function type for digest computation.
type DigestFn func(method *Method, key ctypes.Bytes, data ctypes.Bytes) (ctypes.Bytes, error)

// ValidateFn is the function type for digest validation.
type ValidateFn func(method *Method, key ctypes.Bytes, signature ctypes.Bytes, data ctypes.Bytes) (bool, error)
