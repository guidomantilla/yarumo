package hashes

import (
	_ "crypto/sha256"
	_ "crypto/sha3"
	_ "crypto/sha512"
	"sync"

	_ "golang.org/x/crypto/blake2b"
)

var methods = map[string]Method{
	SHA256.name:      *SHA256,
	SHA512.name:      *SHA512,
	SHA3_256.name:    *SHA3_256,
	SHA3_512.name:    *SHA3_512,
	BLAKE2b_256.name: *BLAKE2b_256,
	BLAKE2b_512.name: *BLAKE2b_512,
}

var lock = new(sync.RWMutex)

// Register adds a hash method to the registry.
func Register(method Method) {
	lock.Lock()
	defer lock.Unlock()

	methods[method.name] = method
}

// Get retrieves a hash method by name from the registry.
func Get(name string) (*Method, error) {
	lock.RLock()
	defer lock.RUnlock()

	alg, ok := methods[name]
	if !ok {
		return nil, ErrAlgorithmNotSupported(name)
	}

	return &alg, nil
}

// Supported returns all registered hash methods.
func Supported() []Method {
	lock.RLock()
	defer lock.RUnlock()

	list := make([]Method, 0, len(methods))
	for _, method := range methods {
		list = append(list, method)
	}

	return list
}
