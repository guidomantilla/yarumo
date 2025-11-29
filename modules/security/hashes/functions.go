package hashes

import (
	"crypto"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
)

func Hash(hash crypto.Hash, data types.Bytes) types.Bytes {
	assert.True(hash.Available(), "hash function not available")

	h := hash.New()
	_, _ = h.Write(data)
	return h.Sum(nil)
}
