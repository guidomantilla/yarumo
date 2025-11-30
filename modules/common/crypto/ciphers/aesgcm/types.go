package aesgcm

import (
	"crypto/cipher"

	"github.com/guidomantilla/yarumo/common/crypto/ciphers/aesgcm/internal"
	"github.com/guidomantilla/yarumo/common/types"
)

var (
	_ AeadFn    = internal.AESGCM
	_ AeadFn    = internal.ChaCha20Poly1305
	_ EncryptFn = encrypt
	_ DecryptFn = decrypt
)

type AeadFn func(key types.Bytes, nonceSize int) (cipher.AEAD, error)

type KeyFn func(method *Method) types.Bytes

type EncryptFn func(method *Method, key types.Bytes, data types.Bytes, aad types.Bytes) (types.Bytes, error)

type DecryptFn func(method *Method, key, ciphered types.Bytes, aad types.Bytes) (types.Bytes, error)
