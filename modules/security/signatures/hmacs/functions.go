package hmacs

import (
	"crypto"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"hash"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
)

func Sign(hash crypto.Hash, key types.Bytes, data types.Bytes) types.Bytes {
	assert.True(hash.Available(), "hash function not available")

	h := hmac.New(hash.New, key)
	_, _ = h.Write(data)
	return h.Sum(nil)
}

func Verify(hash crypto.Hash, key types.Bytes, signature types.Bytes, data types.Bytes) bool {
	assert.True(hash.Available(), "hash function not available")
	calculated := Sign(hash, key, data)
	return hmac.Equal(signature, calculated)
}

// internal use only

func sha_new256() hash.Hash {
	return sha256.New()
}

func sha_new512() hash.Hash {
	return sha512.New()
}
