package hmacs

import (
	_ "crypto/sha256"
	_ "crypto/sha512"
	"sync"
)

var methods = map[string]Method{
	HMAC_with_SHA256.name: *HMAC_with_SHA256,
	HMAC_with_SHA512.name: *HMAC_with_SHA512,
}

var lock = new(sync.RWMutex)

// Register adds an HMAC method to the registry.
func Register(method Method) {
	lock.Lock()
	defer lock.Unlock()

	methods[method.name] = method
}

// Get retrieves an HMAC method by name from the registry.
func Get(name string) (*Method, error) {
	lock.RLock()
	defer lock.RUnlock()

	alg, ok := methods[name]
	if !ok {
		return nil, ErrAlgorithmNotSupported(name)
	}

	return &alg, nil
}

// Supported returns all registered HMAC methods.
func Supported() []Method {
	lock.RLock()
	defer lock.RUnlock()

	list := make([]Method, 0, len(methods))
	for _, method := range methods {
		list = append(list, method)
	}

	return list
}
