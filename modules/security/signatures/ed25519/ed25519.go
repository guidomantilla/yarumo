package ed25519

import (
	"crypto/ed25519"
	"crypto/rand"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	Ed25519 = NewMethod("Ed25519")
)

type Method struct {
	name string
}

func NewMethod(name string) *Method {
	return &Method{name: name}
}

func (m *Method) Name() string {
	assert.NotNil(m, "method is nil")
	return m.name
}

func (m *Method) GenerateKey() (ed25519.PrivateKey, error) {
	assert.NotNil(m, "method is nil")
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (m *Method) Sign(key *ed25519.PrivateKey, data types.Bytes) (types.Bytes, error) {
	assert.NotNil(m, "method is nil")
	return Sign(m, key, data)
}

func (m *Method) Verify(key *ed25519.PublicKey, signature, data types.Bytes) (bool, error) {
	assert.NotNil(m, "method is nil")
	return Verify(m, key, signature, data)
}
