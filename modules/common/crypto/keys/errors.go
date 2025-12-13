package keys

import "errors"

var (
	ErrMethodIsNil      = errors.New("hkdf: method is nil")
	ErrHashNotAvailable = errors.New("hkdf: hash function not available")
	ErrLengthTooLarge   = errors.New("hkdf: output length too large")
)
