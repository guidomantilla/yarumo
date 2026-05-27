package kdfs

import (
	_ "crypto/sha256"
	_ "crypto/sha512"
	"sync"
)

var methods = map[string]Method{
	HKDF_with_SHA256.name:   *HKDF_with_SHA256,
	HKDF_with_SHA384.name:   *HKDF_with_SHA384,
	HKDF_with_SHA512.name:   *HKDF_with_SHA512,
	PBKDF2_with_SHA256.name: *PBKDF2_with_SHA256,
	PBKDF2_with_SHA512.name: *PBKDF2_with_SHA512,
	Scrypt_KDF.name:         *Scrypt_KDF,
}

var lock = new(sync.RWMutex)

// Register adds a KDF method to the registry.
func Register(method Method) {
	lock.Lock()
	defer lock.Unlock()

	methods[method.name] = method
}

// Get retrieves a KDF method by name from the registry. The returned pointer
// references a snapshot taken at lookup time; subsequent Register calls
// do not affect previously returned pointers. Callers that need fresh
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

// Supported returns all registered KDF methods.
func Supported() []Method {
	lock.RLock()
	defer lock.RUnlock()

	list := make([]Method, 0, len(methods))
	for _, method := range methods {
		list = append(list, method)
	}

	return list
}
