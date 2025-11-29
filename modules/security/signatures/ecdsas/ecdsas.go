package ecdsas

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"hash"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	ECDSA_with_SHA256_over_P256 = NewMethod("ECDSA_with_SHA256_over_P256", crypto.SHA256, sha_new256, 32, elliptic.P256())
	ECDSA_with_SHA512_over_P521 = NewMethod("ECDSA_with_SHA512_over_P521", crypto.SHA512, sha_new512, 66, elliptic.P521())
)

type Method struct {
	name    string
	kind    crypto.Hash
	fn      func() hash.Hash
	keySize int
	curve   elliptic.Curve
}

func NewMethod(name string, kind crypto.Hash, fn func() hash.Hash, keySize int, curve elliptic.Curve) *Method {
	return &Method{
		name:    name,
		kind:    kind,
		fn:      fn,
		keySize: keySize,
		curve:   curve,
	}
}

func (m *Method) Name() string {
	assert.NotNil(m, "method is nil")
	return m.name
}

func (m *Method) Key() (*ecdsa.PrivateKey, error) {
	assert.NotNil(m, "method is nil")
	return ecdsa.GenerateKey(m.curve, rand.Reader)
}

func (m *Method) Sign(key *ecdsa.PrivateKey, data types.Bytes) (types.Bytes, error) {
	assert.NotNil(m, "method is nil")
	return Sign(m, key, data)
}

func (m *Method) Verify(key types.Bytes, signature types.Bytes, data types.Bytes) bool {
	assert.NotNil(m, "method is nil")
	return Verify(m.kind, key, signature, data)
}
