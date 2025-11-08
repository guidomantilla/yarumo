package cryptos

type chaCha20Cipher struct {
	key []byte
}

func NewChaCha20Cipher(opts ...ChaCha20CipherOption) Cipher {
	options := NewChaCha20CipherOptions(opts...)
	return &chaCha20Cipher{
		key: options.key,
	}
}

func (c *chaCha20Cipher) Encrypt(plainText []byte) ([]byte, error) {
	return ChaCha20Encrypt(c.key, plainText)
}

func (c *chaCha20Cipher) Decrypt(cipherText []byte) ([]byte, error) {
	return ChaCha20Decrypt(c.key, cipherText)
}
