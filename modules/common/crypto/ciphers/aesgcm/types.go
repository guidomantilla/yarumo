package aesgcm

import "crypto/cipher"

type AeadFn func(key []byte) (cipher.AEAD, error)
