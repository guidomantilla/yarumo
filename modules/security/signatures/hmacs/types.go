package hmacs

import (
	"crypto"

	"github.com/guidomantilla/yarumo/common/types"
)

var (
	_ SignFn   = Sign
	_ VerifyFn = Verify
)

type SignFn func(hash crypto.Hash, key types.Bytes, data types.Bytes) types.Bytes
type VerifyFn func(hash crypto.Hash, key types.Bytes, signature types.Bytes, data types.Bytes) bool
