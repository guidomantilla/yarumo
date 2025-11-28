package macs

import "errors"

var (
	ErrAlgorithmNotSupported = errors.New("algorithm not supported")
	ErrDataEmpty             = errors.New("data is empty")
	ErrKeySizeInvalid        = errors.New("key size invalid")
)
