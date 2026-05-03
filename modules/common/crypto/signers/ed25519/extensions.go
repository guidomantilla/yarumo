package ed25519

import (
	"sync"
)

var methods = map[string]Method{
	Ed25519.name: *Ed25519,
}

var lock = new(sync.RWMutex)

// Register adds an Ed25519 method to the registry.
func Register(method Method) {
	lock.Lock()
	defer lock.Unlock()

	methods[method.name] = method
}

// Get retrieves an Ed25519 method by name from the registry.
func Get(name string) (*Method, error) {
	lock.RLock()
	defer lock.RUnlock()

	alg, ok := methods[name]
	if !ok {
		return nil, ErrAlgorithmNotSupported(name)
	}

	return &alg, nil
}

// Supported returns all registered Ed25519 methods.
func Supported() []Method {
	lock.RLock()
	defer lock.RUnlock()

	list := make([]Method, 0, len(methods))
	for _, method := range methods {
		list = append(list, method)
	}

	return list
}
