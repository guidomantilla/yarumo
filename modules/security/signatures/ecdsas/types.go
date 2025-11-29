package ecdsas

import (
	"crypto"
	"crypto/ecdsa"

	"github.com/guidomantilla/yarumo/common/types"
)

var (
	_ SignFn   = ECDSA_P256_SHA256
	_ VerifyFn = ECDSA_P521_SHA512
)

type SignFn func(key *ecdsa.PrivateKey, data types.Bytes) (types.Bytes, error)

type VerifyFn func(key *ecdsa.PublicKey, signature types.Bytes, data types.Bytes) bool
