package hmacs

import (
	"crypto"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/random"
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	HMAC_with_SHA256 = NewMethod("HMAC_with_SHA256", crypto.SHA256, 32)
	HMAC_with_SHA512 = NewMethod("HMAC_with_SHA512", crypto.SHA512, 64)
)

type Method struct {
	name    string
	kind    crypto.Hash
	keySize int
}

func NewMethod(name string, kind crypto.Hash, keySize int) *Method {
	return &Method{
		name:    name,
		kind:    kind,
		keySize: keySize,
	}
}

func (m *Method) Name() string {
	assert.NotNil(m, "method is nil")
	return m.name
}

func (m *Method) GenerateKey() types.Bytes {
	assert.NotNil(m, "method is nil")
	return random.Key(m.keySize)
}

func (m *Method) Digest(key types.Bytes, data types.Bytes) types.Bytes {
	assert.NotNil(m, "method is nil")
	return Digest(m.kind, key, data)
}

func (m *Method) Validate(key types.Bytes, signature types.Bytes, data types.Bytes) bool {
	assert.NotNil(m, "method is nil")
	return Validate(m.kind, key, signature, data)
}
