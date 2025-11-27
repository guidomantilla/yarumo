package hashes

var (
	_ HashFn = SHA256
	_ HashFn = SHA3_256
	_ HashFn = BLAKE2b_256
	_ HashFn = SHA512
	_ HashFn = SHA3_512
	_ HashFn = BLAKE2b_512
)

type HashFn func(data []byte) []byte
