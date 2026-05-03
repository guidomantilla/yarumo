package rsaoaep

import (
	"sync"
)

var methods = map[string]Method{
	RSA_OAEP_SHA256.name: *RSA_OAEP_SHA256,
	RSA_OAEP_SHA512.name: *RSA_OAEP_SHA512,
}

var lock = new(sync.RWMutex)

// Register adds an RSA-OAEP method to the registry.
func Register(method Method) {
	lock.Lock()
	defer lock.Unlock()

	methods[method.name] = method
}

// Get retrieves an RSA-OAEP method by name from the registry.
func Get(name string) (*Method, error) {
	lock.RLock()
	defer lock.RUnlock()

	alg, ok := methods[name]
	if !ok {
		return nil, ErrAlgorithmNotSupported(name)
	}

	return &alg, nil
}

// Supported returns all registered RSA-OAEP methods.
func Supported() []Method {
	lock.RLock()
	defer lock.RUnlock()

	list := make([]Method, 0, len(methods))
	for _, method := range methods {
		list = append(list, method)
	}

	return list
}
