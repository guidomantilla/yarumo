package cryptos

var (
	_ CipherFn = Encrypt
	_ CipherFn = Decrypt
)

type CipherFn func(key []byte, plaintext []byte) ([]byte, error)

//

var (
	_ Cipher = (*aesCipher)(nil)
)

type Cipher interface {
	Encrypt(plainText []byte) ([]byte, error)
	Decrypt(cipherText []byte) ([]byte, error)
}
