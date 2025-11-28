package types

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/guidomantilla/yarumo/common/assert"
)

type Bytes []byte

func (b Bytes) ToHex() string {
	assert.NotNil(b, "[]byte is nil")
	return hex.EncodeToString(b)
}

func (b Bytes) ToBase64Std() string {
	assert.NotNil(b, "[]byte is nil")
	return base64.StdEncoding.EncodeToString(b)
}

func (b Bytes) ToBase64RawStd() string {
	assert.NotNil(b, "[]byte is nil")
	return base64.RawStdEncoding.EncodeToString(b)
}

func (b Bytes) ToBase64Url() string {
	assert.NotNil(b, "[]byte is nil")
	return base64.URLEncoding.EncodeToString(b)
}

func (b Bytes) ToBase64RawUrl() string {
	assert.NotNil(b, "[]byte is nil")
	return base64.RawURLEncoding.EncodeToString(b)
}
