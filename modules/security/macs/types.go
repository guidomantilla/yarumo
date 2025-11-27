package macs

var (
	_ MacFn = HMAC_SHA256
	_ MacFn = HMAC_SHA3_256
	_ MacFn = BLAKE2b_256_MAC
	_ MacFn = HMAC_SHA512
	_ MacFn = HMAC_SHA3_512
	_ MacFn = BLAKE2b_512_MAC
)

type MacFn func(key []byte, data []byte) []byte
