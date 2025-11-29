package hashes

import (
	"crypto"
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"hash"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"
	"golang.org/x/crypto/blake2b"
)

func Hash(hash crypto.Hash, data types.Bytes) types.Bytes {
	assert.True(hash.Available(), "hash function not available")

	h := hash.New()
	_, _ = h.Write(data)
	return h.Sum(nil)
}

// internal use only

func sha_new256() hash.Hash {
	return sha256.New()
}

func sha_new512() hash.Hash {
	return sha512.New()
}

func sha3_new256() hash.Hash {
	return sha3.New256()
}

func sha3_new512() hash.Hash {
	return sha3.New512()
}

func blake2b_new256() hash.Hash {
	h, _ := blake2b.New256(nil)
	return h
}

func blake2b_new512() hash.Hash {
	h, _ := blake2b.New512(nil)
	return h
}
