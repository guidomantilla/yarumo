package hmacs

import (
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	_ DigestFn   = Digest
	_ ValidateFn = Validate
)

type DigestFn func(method *Method, key types.Bytes, data types.Bytes) (types.Bytes, error)

type ValidateFn func(method *Method, key types.Bytes, signature types.Bytes, data types.Bytes) (bool, error)
