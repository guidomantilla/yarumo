package hashes

import (
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"

	"golang.org/x/crypto/blake2b"
)

func SHA256(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}

func SHA3_256(data []byte) []byte {
	d := sha3.New256()
	_, _ = d.Write(data)
	return d.Sum(nil)
}

func SHA512(data []byte) []byte {
	sum := sha512.Sum512(data)
	return sum[:]
}

func SHA3_512(data []byte) []byte {
	d := sha3.New512()
	_, _ = d.Write(data)
	return d.Sum(nil)
}

func BLAKE2b_512(data []byte) []byte {
	d, _ := blake2b.New512(nil)
	d.Write(data)
	return d.Sum(nil)
}
