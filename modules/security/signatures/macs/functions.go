package macs

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"

	"github.com/guidomantilla/yarumo/security/keys"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"
)

func Key(name Name) ([]byte, error) {
	alg, err := Get(name)
	if err != nil {
		return nil, err
	}
	return keys.Key(alg.KeySize), nil
}

// 256

func HMAC_SHA256(key []byte, data []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, ErrKeySizeInvalid
	}
	if len(data) == 0 {
		return nil, ErrDataEmpty
	}

	if len(key) != 32 {
		return nil, ErrKeySizeInvalid
	}

	m := hmac.New(sha256.New, key)
	m.Write(data)
	return m.Sum(nil), nil
}

func HMAC_SHA3_256(key []byte, data []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, ErrKeySizeInvalid
	}
	if len(data) == 0 {
		return nil, ErrDataEmpty
	}

	if len(key) != 32 {
		return nil, ErrKeySizeInvalid
	}

	m := hmac.New(sha3.New256, key)
	m.Write(data)
	return m.Sum(nil), nil
}

func BLAKE2b_256_MAC(key []byte, data []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, ErrKeySizeInvalid
	}
	if len(data) == 0 {
		return nil, ErrDataEmpty
	}

	if len(key) != 32 {
		return nil, ErrKeySizeInvalid
	}

	d, err := blake2b.New256(key)
	if err != nil {
		return nil, err
	}
	d.Write(data)
	return d.Sum(nil), nil
}

// 512

func HMAC_SHA512(key []byte, data []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, ErrKeySizeInvalid
	}
	if len(data) == 0 {
		return nil, ErrDataEmpty
	}

	if len(key) != 64 {
		return nil, ErrKeySizeInvalid
	}

	m := hmac.New(sha512.New, key)
	m.Write(data)
	return m.Sum(nil), nil
}

func HMAC_SHA3_512(key []byte, data []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, ErrKeySizeInvalid
	}
	if len(data) == 0 {
		return nil, ErrDataEmpty
	}

	if len(key) != 64 {
		return nil, ErrKeySizeInvalid
	}

	m := hmac.New(sha3.New512, key)
	m.Write(data)
	return m.Sum(nil), nil
}

func BLAKE2b_512_MAC(key []byte, data []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, ErrKeySizeInvalid
	}
	if len(data) == 0 {
		return nil, ErrDataEmpty
	}

	if len(key) != 64 {
		return nil, ErrKeySizeInvalid
	}

	d, err := blake2b.New512(key)
	if err != nil {
		return nil, err
	}
	d.Write(data)
	return d.Sum(nil), nil
}

func Equal(a []byte, b []byte) bool {
	return hmac.Equal(a, b)
}
