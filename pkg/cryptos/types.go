package cryptos

var (
	_ KeyFn    = Key
	_ CipherFn = Encrypt
	_ CipherFn = Decrypt
)

type KeyFn func(size int) (*string, error)

type CipherFn func(key []byte, plaintext []byte) ([]byte, error)

//

var (
	_ Cipher = (*aesCipher)(nil)
)

type Cipher interface {
	Encrypt(plainText []byte) ([]byte, error)
	Decrypt(cipherText []byte) ([]byte, error)
}
