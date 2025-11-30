package ed25519

import (
	"crypto/ed25519"

	"github.com/guidomantilla/yarumo/common/types"
)

var (
	_ KeyFn    = key
	_ SignFn   = sign
	_ VerifyFn = verify
)

type KeyFn func() (ed25519.PublicKey, ed25519.PrivateKey, error)

type SignFn func(method *Method, key *ed25519.PrivateKey, data types.Bytes) (types.Bytes, error)

type VerifyFn func(method *Method, key *ed25519.PublicKey, signature types.Bytes, data types.Bytes) (bool, error)
