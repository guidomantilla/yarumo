package rsapss

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
	"github.com/guidomantilla/yarumo/common/utils"
)

var (
	RSASSA_PSS_SHA256 = NewMethod("RSASSA_PSS_SHA256", crypto.SHA256, rsa.PSSSaltLengthEqualsHash, 2048, 3072, 4096)
	RSASSA_PSS_SHA512 = NewMethod("RSASSA_PSS_SHA512", crypto.SHA512, rsa.PSSSaltLengthEqualsHash, 3072, 4096)
)

type Method struct {
	name            string
	kind            crypto.Hash
	saltLength      int
	allowedKeySizes []int
}

func NewMethod(name string, kind crypto.Hash, saltLength int, allowedKeySizes ...int) *Method {
	return &Method{
		name:            name,
		kind:            kind,
		saltLength:      saltLength,
		allowedKeySizes: allowedKeySizes,
	}
}
func (m *Method) Name() string {
	assert.NotNil(m, "method is nil")
	return m.name
}

func (m *Method) GenerateKey(size int) (*rsa.PrivateKey, error) {
	assert.NotNil(m, "method is nil")
	if utils.NotIn(size, m.allowedKeySizes...) {
		return nil, ErrKeySizeNotAllowed
	}
	return rsa.GenerateKey(rand.Reader, size)
}

func (m *Method) Sign(key *rsa.PrivateKey, data types.Bytes) (types.Bytes, error) {
	assert.NotNil(m, "method is nil")
	return Sign(m, key, data)
}

func (m *Method) Verify(key *rsa.PublicKey, signature types.Bytes, data types.Bytes) (bool, error) {
	assert.NotNil(m, "method is nil")
	return Verify(m, key, signature, data)
}
