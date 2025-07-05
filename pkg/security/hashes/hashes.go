package hashes

import "fmt"

const (
	Sha256      = "SHA256"
	Sha3_256    = "SHA3_256"
	Sha512      = "SHA512"
	Sha3_512    = "SHA3_512"
	Blake2b_512 = "BLAKE2b_512"
)

func GetByName(name string) (HashFn, error) {
	switch name {
	case Sha256:
		return SHA256, nil
	case Sha3_256:
		return SHA3_256, nil
	case Sha512:
		return SHA512, nil
	case Sha3_512:
		return SHA3_512, nil
	case Blake2b_512:
		return BLAKE2b_512, nil
	default:
		return nil, fmt.Errorf("hash function %s not found", name)
	}
}
