package internal

import (
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"hash"

	"golang.org/x/crypto/blake2b"
)

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
