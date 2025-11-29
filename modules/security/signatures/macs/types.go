package macs

import (
	"crypto"

	"github.com/guidomantilla/yarumo/common/types"
)

type Name string

const (
	HMAC_with_SHA_256  Name = "HMAC-SHA256"
	HMAC_with_SHA3_256 Name = "HMAC-SHA3-256"
	MB2b_256           Name = "BLAKE2b-256-MAC"
	HS_512             Name = "HMAC-SHA512"
	HS3_512            Name = "HMAC-SHA3-512"
	MB2b_512           Name = "BLAKE2b-512-MAC"
)

var (
	_ Fn = HMAC_SHA256
	_ Fn = HMAC_SHA3_256
	_ Fn = BLAKE2b_256_MAC
	_ Fn = HMAC_SHA512
	_ Fn = HMAC_SHA3_512
	_ Fn = BLAKE2b_512_MAC
)

type Fn func(key types.Bytes, data types.Bytes) (types.Bytes, error)

type Algorithm struct {
	Name    Name        `json:"name"`
	Alias   []Name      `json:"alias"`
	HashFn  crypto.Hash `json:"-"`
	KeySize int         `json:"key-size"`
}
