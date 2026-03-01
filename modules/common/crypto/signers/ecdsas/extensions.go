package ecdsas

import (
	"sync"
)

var methods = map[string]Method{
	ECDSA_with_SHA256_over_P256.name: *ECDSA_with_SHA256_over_P256,
	ECDSA_with_SHA512_over_P521.name: *ECDSA_with_SHA512_over_P521,
}

var lock = new(sync.RWMutex)

// Register adds an ECDSA method to the registry.
func Register(method Method) {
	lock.Lock()
	defer lock.Unlock()

	methods[method.name] = method
}

// Get retrieves an ECDSA method by name from the registry.
func Get(name string) (*Method, error) {
	lock.RLock()
	defer lock.RUnlock()

	alg, ok := methods[name]
	if !ok {
		return nil, ErrAlgorithmNotSupported(name)
	}

	return &alg, nil
}

// Supported returns all registered ECDSA methods.
func Supported() []Method {
	lock.RLock()
	defer lock.RUnlock()

	list := make([]Method, 0, len(methods))
	for _, method := range methods {
		list = append(list, method)
	}

	return list
}
