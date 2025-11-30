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

// Register adds a new HMAC method to the registry.
func Register(method Method) {
	lock.Lock()
	defer lock.Unlock()

	methods[method.name] = method
}

// Get returns the HMAC method with the given name.
func Get(name string) (*Method, error) {
	lock.Lock()
	defer lock.Unlock()

	alg, ok := methods[name]
	if !ok {
		return nil, ErrAlgorithmNotSupported(name)
	}
	return &alg, nil
}

// Supported returns a list of all supported HMAC methods.
func Supported() []Method {
	lock.Lock()
	defer lock.Unlock()

	var list []Method
	for _, method := range methods {
		list = append(list, method)
	}
	return list
}
