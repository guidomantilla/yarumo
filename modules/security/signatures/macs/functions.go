package macs

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"

	"github.com/guidomantilla/yarumo/common/types"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"
)

// 256

func HMAC_SHA256(key types.Bytes, data types.Bytes) (types.Bytes, error) {
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

func HMAC_SHA3_256(key types.Bytes, data types.Bytes) (types.Bytes, error) {
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

func BLAKE2b_256_MAC(key types.Bytes, data types.Bytes) (types.Bytes, error) {
	if len(key) == 0 {
		return nil, ErrKeySizeInvalid
	}
	if len(data) == 0 {
		return nil, ErrDataEmpty
	}

	if len(key) != 32 {
		return nil, ErrKeySizeInvalid
	}

    // With enforced key size, blake2b.New256 will not return an error.
    d, _ := blake2b.New256(key)
    d.Write(data)
    return d.Sum(nil), nil
}

// 512

func HMAC_SHA512(key types.Bytes, data types.Bytes) (types.Bytes, error) {
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

func HMAC_SHA3_512(key types.Bytes, data types.Bytes) (types.Bytes, error) {
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

func BLAKE2b_512_MAC(key types.Bytes, data types.Bytes) (types.Bytes, error) {
	if len(key) == 0 {
		return nil, ErrKeySizeInvalid
	}
	if len(data) == 0 {
		return nil, ErrDataEmpty
	}

	if len(key) != 64 {
		return nil, ErrKeySizeInvalid
	}

    // With enforced key size, blake2b.New512 will not return an error.
    d, _ := blake2b.New512(key)
    d.Write(data)
    return d.Sum(nil), nil
}

func Equal(a types.Bytes, b types.Bytes) bool {
	return hmac.Equal(a, b)
}

func NotEqual(a types.Bytes, b types.Bytes) bool {
	return !Equal(a, b)
}
