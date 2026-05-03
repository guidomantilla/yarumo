package passwords

import (
	"sync"
)

var methods = map[string]Method{
	Argon2.name: *Argon2,
	Bcrypt.name: *Bcrypt,
	Pbkdf2.name: *Pbkdf2,
	Scrypt.name: *Scrypt,
}

var lock = new(sync.RWMutex)

// Register adds a method to the registry.
func Register(method Method) {
	lock.Lock()
	defer lock.Unlock()
	methods[method.name] = method
}

// Get retrieves a method by name from the registry.
func Get(name string) (*Method, error) {
	lock.RLock()
	defer lock.RUnlock()
	alg, ok := methods[name]
	if !ok {
		return nil, ErrAlgorithmNotSupported(name)
	}
	return &alg, nil
}

// Supported returns all registered methods.
func Supported() []Method {
	lock.RLock()
	defer lock.RUnlock()
	list := make([]Method, 0, len(methods))
	for _, method := range methods {
		list = append(list, method)
	}
	return list
}

// ByPrefix returns the method matching the encoded password prefix.
func ByPrefix(encodedPassword string) (*Method, error) {
	lock.RLock()
	defer lock.RUnlock()
	for _, method := range methods {
		if len(encodedPassword) > len(method.prefix) && encodedPassword[:len(method.prefix)] == method.prefix {
			return &method, nil
		}
	}
	return nil, ErrAlgorithmNotSupported(encodedPassword)
}
