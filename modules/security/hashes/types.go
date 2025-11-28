package hashes

import (
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	_ HashFn = SHA256
	_ HashFn = SHA3_256
	_ HashFn = BLAKE2b_256
	_ HashFn = SHA512
	_ HashFn = SHA3_512
	_ HashFn = BLAKE2b_512
)

type HashFn func(data types.Bytes) types.Bytes
