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
