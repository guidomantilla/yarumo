package hashes

import (
	"crypto"

	"github.com/guidomantilla/yarumo/common/types"
)

var (
	_ Fn = Hash
)

type Fn func(hash crypto.Hash, data types.Bytes) types.Bytes
