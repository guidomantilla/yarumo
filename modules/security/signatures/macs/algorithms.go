package macs

import "crypto"

var (
	HMAC_with_SHA_256 = Algorithm{Name: "HMAC + SHA256", HashFn: crypto.SHA256, KeySize: 32, Alias: []Name{}}
)
