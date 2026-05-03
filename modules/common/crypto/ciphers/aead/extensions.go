package aead

import (
	"sync"
)

var methods = map[string]Method{
	AES_128_GCM.name:        *AES_128_GCM,
	AES_256_GCM.name:        *AES_256_GCM,
	CHACHA20_POLY1305.name:  *CHACHA20_POLY1305,
	XCHACHA20_POLY1305.name: *XCHACHA20_POLY1305,
}

var lock = new(sync.RWMutex)

// Register adds an AEAD method to the registry.
func Register(method Method) {
	lock.Lock()
	defer lock.Unlock()

	methods[method.name] = method
}

// Get retrieves an AEAD method by name from the registry.
func Get(name string) (*Method, error) {
	lock.RLock()
	defer lock.RUnlock()

	alg, ok := methods[name]
	if !ok {
		return nil, ErrAlgorithmNotSupported(name)
	}

	return &alg, nil
}

// Supported returns all registered AEAD methods.
func Supported() []Method {
	lock.RLock()
	defer lock.RUnlock()

	list := make([]Method, 0, len(methods))
	for _, method := range methods {
		list = append(list, method)
	}

	return list
}
