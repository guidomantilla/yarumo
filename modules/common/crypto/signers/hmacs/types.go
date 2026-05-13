// Package hmacs provides HMAC (Hash-based Message Authentication Code) key
// generation, digest computation, and validation using pluggable hash functions
// and a thread-safe name-based registry.
//
// # Config-driven algorithm selection
//
// *Method implements encoding.TextMarshaler / encoding.TextUnmarshaler.
// MarshalText emits the registered algorithm name; UnmarshalText resolves a
// name against the package registry (via Get) and overwrites the receiver.
// This makes Method directly compatible with libraries that honor the
// encoding interfaces — including encoding/json, viper, kong, and koanf —
// so deployments can load HMAC algorithm choice from YAML/JSON/TOML config.
//
// Caveat: UnmarshalText resolves against whatever the registry contains at
// the time of the call. Custom methods registered via Register after config
// load will not resolve here; callers that need late-bound lookup should
// call Get(name) directly.
//
// # Recommended entry point for string-named algorithms
//
// Digest(name, key, data) and Validate(name, key, digest, data) are the
// recommended top-level helpers for callers that load the algorithm name
// from config. They each perform a single Get and forward to the
// corresponding Method operation, returning the package's domain error
// when the name is not registered.
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
