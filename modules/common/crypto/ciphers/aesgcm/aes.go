package aesgcm

type Method struct {
	name      string
	keySize   int
	nonceSize int
	aeadFn    AeadFn
}

func NewMethod(name string, keySize, nonceSize int, aeadFn AeadFn) *Method {
	return &Method{
		name:      name,
		keySize:   keySize,
		nonceSize: nonceSize,
		aeadFn:    aeadFn,
	}
}
