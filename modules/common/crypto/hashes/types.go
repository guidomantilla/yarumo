// Package hashes provides cryptographic hash digest computation using Go's
// standard crypto.Hash primitives. It wraps the standard hash.New / Write / Sum
// pattern into a single Hash call and a Method descriptor with a thread-safe
// name-based registry.
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
type HashFn func(hash crypto.Hash, data ctypes.Bytes) ctypes.Bytes
