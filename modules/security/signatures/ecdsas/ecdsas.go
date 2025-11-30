package ecdsas

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	ECDSA_with_SHA256_over_P256 = NewMethod("ECDSA_with_SHA256_over_P256", crypto.SHA256, 32, elliptic.P256())
	ECDSA_with_SHA512_over_P521 = NewMethod("ECDSA_with_SHA512_over_P521", crypto.SHA512, 66, elliptic.P521())
)

type Method struct {
	name    string
	kind    crypto.Hash
	keySize int
	curve   elliptic.Curve
}

func NewMethod(name string, kind crypto.Hash, keySize int, curve elliptic.Curve) *Method {
	return &Method{
		name:    name,
		kind:    kind,
		keySize: keySize,
		curve:   curve,
	}
}

func (m *Method) Name() string {
	assert.NotNil(m, "method is nil")
	return m.name
}

func (m *Method) GenerateKey() (*ecdsa.PrivateKey, error) {
	assert.NotNil(m, "method is nil")
	return ecdsa.GenerateKey(m.curve, rand.Reader)
}

func (m *Method) Sign(key *ecdsa.PrivateKey, data types.Bytes, format Format) (types.Bytes, error) {
	assert.NotNil(m, "method is nil")
	return Sign(m, key, data, format)
}

func (m *Method) Verify(key *ecdsa.PublicKey, signature types.Bytes, data types.Bytes, format Format) (bool, error) {
	assert.NotNil(m, "method is nil")
	return Verify(m, key, signature, data, format)
}
