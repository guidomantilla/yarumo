package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"

	"github.com/guidomantilla/yarumo/common/types"
)

type Name string

const (
	ES256 Name = "ECDSA-P256-SHA256"
	ES512 Name = "ECDSA-P521-SHA512"
)

var (
	_ EcdsaFn = ECDSA_P256_SHA256
	_ EcdsaFn = ECDSA_P521_SHA512
)

type EcdsaFn func(key *ecdsa.PrivateKey, data types.Bytes) (types.Bytes, error)

type Algorithm struct {
	Name    Name           `json:"name"`
	Alias   Name           `json:"alias"`
	Fn      EcdsaFn        `json:"-"`
	KeySize int            `json:"key-size"`
	Curve   elliptic.Curve `json:"curve"`
}
