package hmacs

import (
	"crypto"

	"github.com/guidomantilla/yarumo/common/types"
)

var (
	_ DigestFn   = Digest
	_ ValidateFn = Validate
)

type DigestFn func(hash crypto.Hash, key types.Bytes, data types.Bytes) types.Bytes
type ValidateFn func(hash crypto.Hash, key types.Bytes, signature types.Bytes, data types.Bytes) bool
