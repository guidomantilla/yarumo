package macs

import (
	"crypto"
	"hash"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/random"
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	HMAC_with_SHA256 = NewMethod("HMAC_with_SHA256", crypto.SHA256, sha_new256, 32)
	HMAC_with_SHA512 = NewMethod("HMAC_with_SHA512", crypto.SHA512, sha_new512, 64)
	//ECDSA_with_SHA256_over_P256 = Algorithm{Name: "ECDSA_with_SHA256_over_P256", Kind: "ECDSA", HashFn: crypto.SHA256, KeySize: 32, Curve: elliptic.P256()}
	//ECDSA_with_SHA512_over_P521 = Algorithm{Name: "ECDSA_with_SHA512_over_P521", Kind: "ECDSA", HashFn: crypto.SHA512, KeySize: 66, Curve: elliptic.P521()}
)

type Method struct {
	name    string
	kind    crypto.Hash
	fn      func() hash.Hash
	keySize int
}

func NewMethod(name string, kind crypto.Hash, fn func() hash.Hash, keySize int) *Method {
	return &Method{
		name:    name,
		kind:    kind,
		fn:      fn,
		keySize: keySize,
	}
}

func (m *Method) Name() string {
	assert.NotNil(m, "method is nil")
	return m.name
}

func (m *Method) Key() types.Bytes {
	assert.NotNil(m, "method is nil")
	return random.Key(m.keySize)
}

func (m *Method) Sign(key types.Bytes, data types.Bytes) types.Bytes {
	assert.NotNil(m, "method is nil")
	return Sign(m.kind, key, data)
}

func (m *Method) Verify(key types.Bytes, signature types.Bytes, data types.Bytes) bool {
	assert.NotNil(m, "method is nil")
	return Verify(m.kind, key, signature, data)
}
