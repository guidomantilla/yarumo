package rsaoaep

import "errors"

var (
	ErrMethodIsNil        = errors.New("method is nil")
	ErrKeyIsNil           = errors.New("key is nil")
	ErrKeyLengthIsInvalid = errors.New("key length is invalid")
	ErrHashNotAvailable   = errors.New("hash function not available")
	ErrEncryptFailed      = errors.New("encryption failed")
	ErrDecryptFailed      = errors.New("decryption failed")
)
