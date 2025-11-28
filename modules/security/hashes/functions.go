package hashes

import (
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"

	"golang.org/x/crypto/blake2b"
)

// 256

func SHA256(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}
	sum := sha256.Sum256(data)
	return sum[:]
}

func SHA3_256(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}
	d := sha3.New256()
	_, _ = d.Write(data)
	return d.Sum(nil)
}

func BLAKE2b_256(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}
	sum := blake2b.Sum256(nil)
	return sum[:]
}

// 512

func SHA512(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}
	sum := sha512.Sum512(data)
	return sum[:]
}

func SHA3_512(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}
	d := sha3.New512()
	_, _ = d.Write(data)
	return d.Sum(nil)
}

func BLAKE2b_512(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}
	sum := blake2b.Sum512(nil)
	return sum[:]
}
