package passwords

import (
	"strings"
	"sync"
)

var methods = map[string]Method{
	Argon2id.name: *Argon2id,
	Argon2i.name:  *Argon2i,
	Bcrypt.name:   *Bcrypt,
	Pbkdf2.name:   *Pbkdf2,
	Scrypt.name:   *Scrypt,
}

var lock = new(sync.RWMutex)

// Register adds a method to the registry.
func Register(method Method) {
	lock.Lock()
	defer lock.Unlock()
	methods[method.name] = method
}

// Get retrieves a method by name from the registry. The returned pointer
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
//
// The legacy {argon2} prefix (used by pre-YA-0030 encodes) is treated as an
// alias of {argon2id} so stored hashes continue to verify against the
// renamed Argon2id method. The {argon2i} prefix routes to Argon2i. Direct
// prefix matches always take precedence over the legacy alias.
func ByPrefix(encodedPassword string) (*Method, error) {
	lock.RLock()
	defer lock.RUnlock()
	for _, method := range methods {
		if len(encodedPassword) > len(method.prefix) && encodedPassword[:len(method.prefix)] == method.prefix {
			return &method, nil
		}
	}
	// Legacy {argon2} prefix — route to Argon2id for backward compatibility
	// with hashes produced before the YA-0030 rename.
	if strings.HasPrefix(encodedPassword, Argon2PrefixKey) && len(encodedPassword) > len(Argon2PrefixKey) {
		if alg, ok := methods[Argon2id.name]; ok {
			return &alg, nil
		}
	}
	return nil, ErrAlgorithmNotSupported(encodedPassword)
}
