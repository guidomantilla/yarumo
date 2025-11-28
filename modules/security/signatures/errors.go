package signatures

import "errors"

var (
	ErrInvalidKeyType   = errors.New("key is of invalid type")
	ErrSignatureInvalid = errors.New("signature is invalid")
)
