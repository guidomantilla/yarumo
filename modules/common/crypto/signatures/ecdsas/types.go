package ecdsas

import (
	"crypto/ecdsa"

	"github.com/guidomantilla/yarumo/common/types"
)

var (
	_ KeyFn    = key
	_ SignFn   = sign
	_ VerifyFn = verify
)

type KeyFn func(method *Method) (*ecdsa.PrivateKey, error)

type SignFn func(method *Method, key *ecdsa.PrivateKey, data types.Bytes, format Format) (types.Bytes, error)

type VerifyFn func(method *Method, key *ecdsa.PublicKey, signature types.Bytes, data types.Bytes, format Format) (bool, error)
