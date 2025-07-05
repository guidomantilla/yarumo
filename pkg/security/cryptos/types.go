package cryptos

var (
	_ KeyFn    = Key
	_ CipherFn = AesEncrypt
	_ CipherFn = AesDecrypt
	_ CipherFn = ChaCha20Encrypt
	_ CipherFn = ChaCha20Decrypt
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
