package macs

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"
)

// 256

func HMAC_SHA256(key []byte, data []byte) []byte {
	if len(key) == 0 {
		return nil
	}
	if len(data) == 0 {
		return nil
	}
	m := hmac.New(sha256.New, key)
	m.Write(data)
	return m.Sum(nil)
}

func HMAC_SHA3_256(key, data []byte) []byte {
	if len(key) == 0 {
		return nil
	}
	if len(data) == 0 {
		return nil
	}
	m := hmac.New(sha3.New256, key)
	m.Write(data)
	return m.Sum(nil)
}

func BLAKE2b_256_MAC(key, data []byte) []byte {
	if len(key) == 0 {
		return nil
	}
	if len(data) == 0 {
		return nil
	}
	d, err := blake2b.New256(key)
	if err != nil {
		return nil
	}
	d.Write(data)
	return d.Sum(nil)
}

// 512

func HMAC_SHA512(key []byte, data []byte) []byte {
	if len(key) == 0 {
		return nil
	}
	if len(data) == 0 {
		return nil
	}
	m := hmac.New(sha512.New, key)
	m.Write(data)
	return m.Sum(nil)
}

func HMAC_SHA3_512(key, data []byte) []byte {
	if len(key) == 0 {
		return nil
	}
	if len(data) == 0 {
		return nil
	}
	m := hmac.New(sha3.New512, key)
	m.Write(data)
	return m.Sum(nil)
}

func BLAKE2b_512_MAC(key, data []byte) []byte {
	if len(key) == 0 {
		return nil
	}
	if len(data) == 0 {
		return nil
	}
	d, err := blake2b.New512(key)
	if err != nil {
		return nil
	}
	d.Write(data)
	return d.Sum(nil)
}
