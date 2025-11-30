package hmacs

import (
	"github.com/guidomantilla/yarumo/common/random"
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	_ KeyFn      = key
	_ DigestFn   = digest
	_ ValidateFn = validate
)

type KeyFn random.KeyFn

type DigestFn func(method *Method, key types.Bytes, data types.Bytes) (types.Bytes, error)

type ValidateFn func(method *Method, key types.Bytes, signature types.Bytes, data types.Bytes) (bool, error)
