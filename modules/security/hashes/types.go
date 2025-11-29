package hashes

import (
	"crypto"

	"github.com/guidomantilla/yarumo/common/types"
)

var (
	_ HashFn = Hash
)

type HashFn func(hash crypto.Hash, data types.Bytes) types.Bytes
