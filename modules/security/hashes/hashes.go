package hashes

import (
	"crypto"
	"hash"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	SHA256      = NewMethod("SHA256", crypto.SHA256, sha_new256)
	SHA512      = NewMethod("SHA512", crypto.SHA512, sha_new512)
	SHA3_256    = NewMethod("SHA3_256", crypto.SHA3_256, sha3_new256)
	SHA3_512    = NewMethod("SHA3_512", crypto.SHA3_512, sha3_new512)
	BLAKE2b_256 = NewMethod("BLAKE2b_256", crypto.BLAKE2b_256, blake2b_new256)
	BLAKE2b_512 = NewMethod("BLAKE2b_512", crypto.BLAKE2b_512, blake2b_new512)
)

type Method struct {
	name string
	kind crypto.Hash
	fn   func() hash.Hash
}

func NewMethod(name string, kind crypto.Hash, fn func() hash.Hash) *Method {
	return &Method{
		name: name,
		kind: kind,
		fn:   fn,
	}
}

func (m *Method) Name() string {
	assert.NotNil(m, "method is nil")
	return m.name
}

func (m *Method) Hash(data types.Bytes) types.Bytes {
	assert.NotNil(m, "method is nil")
	return Hash(m.kind, data)
}
