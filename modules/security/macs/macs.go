package macs

const (
	HmacSha256     = "HMAC_SHA256"
	HmacSha3_256   = "HMAC_SHA3_256"
	Blake2b_256Mac = "BLAKE2b_256_MAC"
	HmacSha512     = "HMAC_SHA512"
	HmacSha3_512   = "HMAC_SHA3_512"
	Blake2b_512Mac = "BLAKE2b_512_MAC"
)

func GetByName(name string) (MacFn, error) {
	switch name {
	case HmacSha256:
		return HMAC_SHA256, nil
	case HmacSha3_256:
		return HMAC_SHA3_256, nil
	case Blake2b_256Mac:
		return BLAKE2b_256_MAC, nil
	case HmacSha512:
		return HMAC_SHA512, nil
	case HmacSha3_512:
		return HMAC_SHA3_512, nil
	case Blake2b_512Mac:
		return BLAKE2b_512_MAC, nil

	default:
		return nil, ErrMacFunctionNotFound(name)
	}
}
