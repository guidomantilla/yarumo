package rsapss

import (
	"crypto/rsa"

	"github.com/guidomantilla/yarumo/common/types"
)

var (
	_ KeyFn    = key
	_ SignFn   = sign
	_ VerifyFn = verify
)

type KeyFn func(bits int) (*rsa.PrivateKey, error)

type SignFn func(method *Method, key *rsa.PrivateKey, data types.Bytes) (types.Bytes, error)

type VerifyFn func(method *Method, key *rsa.PublicKey, signature types.Bytes, data types.Bytes) (bool, error)
