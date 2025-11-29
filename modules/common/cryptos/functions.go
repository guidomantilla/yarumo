package cryptos

import (
	"crypto"
	"crypto/elliptic"
	"crypto/hmac"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/types"

	_ "crypto/sha256"
	_ "crypto/sha3"
	_ "crypto/sha512"

	_ "golang.org/x/crypto/blake2b"
)

type Algorithm struct {
	Name    string         `json:"name"`
	Kind    string         `json:"kind"`
	HashFn  crypto.Hash    `json:"-"`
	KeySize int            `json:"key-size"`
	Curve   elliptic.Curve `json:"curve"`
}

func (a *Algorithm) Hash(data types.Bytes) types.Bytes {
	return Hash(a.HashFn, data)
}

var (
	SHA256                      = Algorithm{Name: "SHA256", Kind: "HASH", HashFn: crypto.SHA256}
	SHA512                      = Algorithm{Name: "SHA512", Kind: "HASH", HashFn: crypto.SHA512}
	SHA3_256                    = Algorithm{Name: "SHA3_256", Kind: "HASH", HashFn: crypto.SHA3_256}
	SHA3_512                    = Algorithm{Name: "SHA3_512", Kind: "HASH", HashFn: crypto.SHA3_512}
	BLAKE2b_256                 = Algorithm{Name: "BLAKE2b_256", Kind: "HASH", HashFn: crypto.BLAKE2b_256}
	BLAKE2b_512                 = Algorithm{Name: "BLAKE2b_512", Kind: "HASH", HashFn: crypto.BLAKE2b_512}
	HMAC_with_SHA256            = Algorithm{Name: "HMAC_with_SHA256", Kind: "HMAC", HashFn: crypto.SHA256}
	HMAC_with_SHA512            = Algorithm{Name: "HMAC_with_SHA512", Kind: "HMAC", HashFn: crypto.SHA512}
	ECDSA_with_SHA256_over_P256 = Algorithm{Name: "ECDSA_with_SHA256_over_P256", Kind: "ECDSA", HashFn: crypto.SHA256, KeySize: 32, Curve: elliptic.P256()}
	ECDSA_with_SHA512_over_P521 = Algorithm{Name: "ECDSA_with_SHA512_over_P521", Kind: "ECDSA", HashFn: crypto.SHA512, KeySize: 66, Curve: elliptic.P521()}
)

// Base Functions

func HmacSign(hash crypto.Hash, key types.Bytes, data types.Bytes) types.Bytes {
	assert.True(hash.Available(), "hash function not available")

	h := hmac.New(hash.New, key)
	_, _ = h.Write(data)
	return h.Sum(nil)
}
