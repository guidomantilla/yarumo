package rsapss

import (
	"sync"
)

var methods = map[string]Method{
	RSASSA_PSS_using_SHA256.name: *RSASSA_PSS_using_SHA256,
	RSASSA_PSS_using_SHA512.name: *RSASSA_PSS_using_SHA512,
}

var lock = new(sync.RWMutex)

// Register adds an RSA-PSS method to the registry.
func Register(method Method) {
	lock.Lock()
	defer lock.Unlock()

	methods[method.name] = method
}

// Get retrieves an RSA-PSS method by name from the registry.
func Get(name string) (*Method, error) {
	lock.RLock()
	defer lock.RUnlock()

	alg, ok := methods[name]
	if !ok {
		return nil, ErrAlgorithmNotSupported(name)
	}

	return &alg, nil
}

// Supported returns all registered RSA-PSS methods.
func Supported() []Method {
	lock.RLock()
	defer lock.RUnlock()

	list := make([]Method, 0, len(methods))
	for _, method := range methods {
		list = append(list, method)
	}

	return list
}
