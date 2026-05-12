// Package hashes provides cryptographic hash digest computation using Go's
// standard crypto.Hash primitives. It wraps the standard hash.New / Write / Sum
// pattern into a single Hash call and a Method descriptor with a thread-safe
// name-based registry.
//
// # Config-driven algorithm selection
//
// *Method implements encoding.TextMarshaler / encoding.TextUnmarshaler.
// MarshalText emits the registered algorithm name; UnmarshalText resolves a
// name against the package registry (via Get) and overwrites the receiver.
// This makes Method directly compatible with libraries that honor the
// encoding interfaces — including encoding/json, viper, kong, and koanf —
// so deployments can load algorithm choice from YAML/JSON/TOML config.
//
// Caveat: UnmarshalText resolves against whatever the registry contains at
// the time of the call. Custom methods registered via Register after config
// load will not resolve here; callers that need late-bound lookup should
// call Get(name) directly.
//
// # Recommended entry point for string-named algorithms
//
// Compute(name, data) is the recommended top-level helper for callers that
// receive the algorithm name as a string (e.g. loaded from config, a request
// header, or a database column). It collapses the standard "Get + Hash"
// boilerplate into a single call and returns the package's domain error
// when the name is not registered. Callers that already hold a *Method
// (predefined or returned by Get) should keep using Method.Hash directly.
//
// # Streaming
//
// Method.NewHasher returns Go's standard hash.Hash so callers can compute
// digests over arbitrary io.Reader / io.Writer sources without materialising
// the entire input in memory. The returned hash.Hash composes directly with
// io.Copy:
//
//	h, err := chashes.SHA256.NewHasher()
//	if err != nil { ... }
//	if _, err := io.Copy(h, src); err != nil { ... }
//	digest := h.Sum(nil)
//
// Use the streaming API for multi-megabyte inputs (file uploads, log
// streams, backup archives) or any time io.Reader composition is natural.
// Method.Hash remains the right choice for short byte buffers already in
// memory.
package hashes

import (
	"crypto"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// Type compliance.
var (
	_ HashFn = Hash
)

// HashFn is the function type for Hash.
type HashFn func(hash crypto.Hash, data ctypes.Bytes) (ctypes.Bytes, error)
