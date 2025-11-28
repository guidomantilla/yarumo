package macs

type Name string

const (
	HS_256   Name = "HMAC-SHA256"
	HS3_256  Name = "HMAC-SHA3-256"
	MB2b_256 Name = "BLAKE2b-256-MAC"
	HS_512   Name = "HMAC-SHA512"
	HS3_512  Name = "HMAC-SHA3-512"
	MB2b_512 Name = "BLAKE2b-512-MAC"
)

var (
	_ MacFn = HMAC_SHA256
	_ MacFn = HMAC_SHA3_256
	_ MacFn = BLAKE2b_256_MAC
	_ MacFn = HMAC_SHA512
	_ MacFn = HMAC_SHA3_512
	_ MacFn = BLAKE2b_512_MAC
)

type MacFn func(key []byte, data []byte) ([]byte, error)

type Algorithm struct {
	Name    Name  `json:"name"`
	Alias   Name  `json:"alias"`
	Fn      MacFn `json:"-"`
	KeySize int   `json:"key-size"`
}
