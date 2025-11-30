package internal

import (
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"hash"

	"golang.org/x/crypto/blake2b"
)

//nolint:unused
func sha_new256() hash.Hash {
	return sha256.New()
}

//nolint:unused
func sha_new512() hash.Hash {
	return sha512.New()
}

//nolint:unused
func sha3_new256() hash.Hash {
	return sha3.New256()
}

//nolint:unused
func sha3_new512() hash.Hash {
	return sha3.New512()
}

//nolint:unused
func blake2b_new256() hash.Hash {
	h, _ := blake2b.New256(nil)
	return h
}

//nolint:unused
func blake2b_new512() hash.Hash {
	h, _ := blake2b.New512(nil)
	return h
}
