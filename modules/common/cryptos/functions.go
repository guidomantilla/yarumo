package cryptos

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"

	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/errs"
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

func Hash(hash crypto.Hash, data types.Bytes) types.Bytes {
	assert.True(hash.Available(), "hash function not available")

	h := hash.New()
	_, _ = h.Write(data)
	return h.Sum(nil)
}

func HmacSign(hash crypto.Hash, key types.Bytes, data types.Bytes) types.Bytes {
	assert.True(hash.Available(), "hash function not available")

	h := hmac.New(hash.New, key)
	_, _ = h.Write(data)
	return h.Sum(nil)
}

func EcdsaSign(hash crypto.Hash, key *ecdsa.PrivateKey, data types.Bytes) (types.Bytes, error) {
	if key == nil {
		return nil, ErrKeyInvalid
	}
	if key.Curve != elliptic.P256() {
		return nil, ErrKeyInvalid
	}
	if len(data) == 0 {
		return nil, ErrDataEmpty
	}

	r, s, err := ecdsa.Sign(rand.Reader, key, Hash(hash, data))
	if err != nil {
		return nil, errs.Wrap(ErrSignFailed, err)
	}

	// We serialize the outputs (r and s) into big-endian byte arrays padded with zeros on the left to make sure the sizes work out.
	// Output must be 2*keyBytes long.
	const keyBytes = 32
	out := make([]byte, 2*keyBytes)

	r.FillBytes(out[0:keyBytes]) // r is assigned to the first half of output.
	s.FillBytes(out[keyBytes:])  // s is assigned to the second half of output.
	return out, nil
}
