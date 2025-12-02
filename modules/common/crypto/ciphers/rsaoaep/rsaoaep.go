package rsaoaep

import (
	"crypto"

	"github.com/guidomantilla/yarumo/common/assert"
)

var (
	RSA_OAEP_SHA256 = NewMethod("RSA-OAEP-SHA256", crypto.SHA256, 2048, 3072, 4096)
	RSA_OAEP_SHA512 = NewMethod("RSA-OAEP-SHA512", crypto.SHA512, 2048, 3072, 4096)
)

type Method struct {
	name            string
	kind            crypto.Hash
	allowedKeySizes []int
}

func NewMethod(name string, kind crypto.Hash, allowedKeySizes ...int) *Method {
	return &Method{
		name:            name,
		kind:            kind,
		allowedKeySizes: allowedKeySizes,
	}
}

func (m *Method) Name() string {
	assert.NotNil(m, "method is nil")
	return m.name
}
