package ecdsa

import "errors"

var (
	ErrAlgorithmNotSupported = errors.New("algorithm not supported")
	ErrDataEmpty             = errors.New("data is empty")
	ErrKeyInvalid            = errors.New("key invalid")
	ErrSignFailed            = errors.New("signing failed")
)
