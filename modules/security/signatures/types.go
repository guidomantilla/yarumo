package signatures

type Signer interface {
	Sign(key any, data []byte) ([]byte, error)
	Verify(key any, signature []byte, data []byte) (bool, error)
}
