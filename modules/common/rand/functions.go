package rand

import (
	"crypto/rand"
)

func Key(size int) []byte {
	key := make([]byte, size)
	_, _ = rand.Reader.Read(key)
	return key
}
