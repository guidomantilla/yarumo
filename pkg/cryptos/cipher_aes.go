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
	ciphertext, err := Encrypt(c.key, plainText)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

func (c *aesCipher) Decrypt(cipherText []byte) ([]byte, error) {
	plaintext, err := Decrypt(c.key, cipherText)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
