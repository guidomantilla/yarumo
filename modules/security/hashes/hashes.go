package hashes

import (
	"crypto"
	"hash"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	SHA256      = NewAlgorithm("SHA256", crypto.SHA256, sha_new256)
	SHA512      = NewAlgorithm("SHA512", crypto.SHA512, sha_new512)
	SHA3_256    = NewAlgorithm("SHA3_256", crypto.SHA3_256, sha3_new256)
	SHA3_512    = NewAlgorithm("SHA3_512", crypto.SHA3_512, sha3_new512)
	BLAKE2b_256 = NewAlgorithm("BLAKE2b_256", crypto.BLAKE2b_256, blake2b_new256)
	BLAKE2b_512 = NewAlgorithm("BLAKE2b_512", crypto.BLAKE2b_512, blake2b_new512)
)

type Algorithm struct {
	name string
	kind crypto.Hash
	fn   func() hash.Hash
}

func NewAlgorithm(name string, kind crypto.Hash, fn func() hash.Hash) *Algorithm {
	return &Algorithm{
		name: name,
		kind: kind,
		fn:   fn,
	}
}

func (a *Algorithm) Name() string {
	assert.NotNil(a, "algorithm is nil")
	return a.name
}

func (a *Algorithm) Hash(data types.Bytes) types.Bytes {
	assert.NotNil(a, "algorithm is nil")
	return Hash(a.kind, data)
}
