package cryptos

import (
	"crypto"
	"crypto/elliptic"
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

var (
	SHA256      = Algorithm{Name: "SHA256", Kind: "HASH", HashFn: crypto.SHA256}
	SHA512      = Algorithm{Name: "SHA512", Kind: "HASH", HashFn: crypto.SHA512}
	SHA3_256    = Algorithm{Name: "SHA3_256", Kind: "HASH", HashFn: crypto.SHA3_256}
	SHA3_512    = Algorithm{Name: "SHA3_512", Kind: "HASH", HashFn: crypto.SHA3_512}
	BLAKE2b_256 = Algorithm{Name: "BLAKE2b_256", Kind: "HASH", HashFn: crypto.BLAKE2b_256}
	BLAKE2b_512 = Algorithm{Name: "BLAKE2b_512", Kind: "HASH", HashFn: crypto.BLAKE2b_512}
)
