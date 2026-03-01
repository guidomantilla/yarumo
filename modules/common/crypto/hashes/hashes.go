package hashes

import (
	"crypto"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// Predefined hash methods registered at package init.
var (
	SHA256      = NewMethod("SHA256", crypto.SHA256)
	SHA512      = NewMethod("SHA512", crypto.SHA512)
	SHA3_256    = NewMethod("SHA3_256", crypto.SHA3_256)
	SHA3_512    = NewMethod("SHA3_512", crypto.SHA3_512)
	BLAKE2b_256 = NewMethod("BLAKE2b_256", crypto.BLAKE2b_256)
	BLAKE2b_512 = NewMethod("BLAKE2b_512", crypto.BLAKE2b_512)
)

// Method holds the configuration for a hash algorithm.
type Method struct {
	name   string
	kind   crypto.Hash
	hashFn HashFn
}

// NewMethod creates a new Method with the given name and hash algorithm.
func NewMethod(name string, kind crypto.Hash, options ...Option) *Method {
	cassert.NotEmpty(name, "name is empty")

	opts := NewOptions(options...)

	return &Method{
		name:   name,
		kind:   kind,
		hashFn: opts.hashFn,
	}
}

// Name returns the method's algorithm name.
func (m *Method) Name() string {
	cassert.NotNil(m, "method is nil")

	return m.name
}

// Hash computes the digest of the given data using this method's hash function.
func (m *Method) Hash(data ctypes.Bytes) ctypes.Bytes {
	cassert.NotNil(m, "method is nil")
	cassert.NotNil(m.hashFn, "method hashFn is nil")

	return m.hashFn(m.kind, data)
}
