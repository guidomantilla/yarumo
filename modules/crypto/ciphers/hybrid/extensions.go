package hybrid

import (
	"sync"
)

var methods = map[string]Method{
	HPKE_X25519_HKDF_SHA256_AES_256_GCM.name: *HPKE_X25519_HKDF_SHA256_AES_256_GCM,
}

var lock = new(sync.RWMutex)

// Register adds a hybrid method to the registry.
func Register(method Method) {
	lock.Lock()
	defer lock.Unlock()

	methods[method.name] = method
}

// Get retrieves a hybrid method by name from the registry. The returned
// pointer references a snapshot taken at lookup time; subsequent Register
// calls do not affect previously returned pointers. Callers that need fresh
// state must call Get again.
func Get(name string) (*Method, error) {
	lock.RLock()
	defer lock.RUnlock()

	alg, ok := methods[name]
	if !ok {
		return nil, ErrAlgorithmNotSupported(name)
	}

	return &alg, nil
}

// Supported returns all registered hybrid methods.
func Supported() []Method {
	lock.RLock()
	defer lock.RUnlock()

	list := make([]Method, 0, len(methods))
	for _, method := range methods {
		list = append(list, method)
	}

	return list
}
