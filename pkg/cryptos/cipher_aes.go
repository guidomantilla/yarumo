package cryptos

type aesCipher struct {
	key []byte
}

func NewAesCipher(opts ...AesCipherOption) Cipher {
	options := NewAesCipherOptions(opts...)
	return &aesCipher{
		key: options.key,
	}
}

func (c *aesCipher) Encrypt(plainText []byte) ([]byte, error) {
	return Encrypt(c.key, plainText)
}

func (c *aesCipher) Decrypt(cipherText []byte) ([]byte, error) {
	return Decrypt(c.key, cipherText)
}
