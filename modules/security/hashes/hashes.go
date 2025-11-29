package hashes

import (
	"crypto"
	"hash"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	SHA256      = NewMethod("SHA256", crypto.SHA256)
	SHA512      = NewMethod("SHA512", crypto.SHA512)
	SHA3_256    = NewMethod("SHA3_256", crypto.SHA3_256)
	SHA3_512    = NewMethod("SHA3_512", crypto.SHA3_512)
	BLAKE2b_256 = NewMethod("BLAKE2b_256", crypto.BLAKE2b_256)
	BLAKE2b_512 = NewMethod("BLAKE2b_512", crypto.BLAKE2b_512)
)

type Method struct {
	name string
	kind crypto.Hash
}

func NewMethod(name string, kind crypto.Hash) *Method {
	return &Method{
		name: name,
		kind: kind,
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
