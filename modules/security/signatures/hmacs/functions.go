package hmacs

import (
	"crypto"
	"crypto/hmac"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
)

func Digest(hash crypto.Hash, key types.Bytes, data types.Bytes) types.Bytes {
	assert.True(hash.Available(), "hash function not available")

	h := hmac.New(hash.New, key)
	_, _ = h.Write(data)
	return h.Sum(nil)
}

func Validate(hash crypto.Hash, key types.Bytes, signature types.Bytes, data types.Bytes) bool {
	assert.True(hash.Available(), "hash function not available")
	calculated := Digest(hash, key, data)
	return hmac.Equal(signature, calculated)
}
