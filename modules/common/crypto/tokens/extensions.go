package tokens

import "sync"

var methods = map[string]Method{
	JWT_HS256.name: *JWT_HS256,
	JWT_HS384.name: *JWT_HS384,
	JWT_HS512.name: *JWT_HS512,

	JWT_RS256.name: *JWT_RS256,
	JWT_RS384.name: *JWT_RS384,
	JWT_RS512.name: *JWT_RS512,

	JWT_PS256.name: *JWT_PS256,
	JWT_PS384.name: *JWT_PS384,
	JWT_PS512.name: *JWT_PS512,

	JWT_ES256.name: *JWT_ES256,
	JWT_ES384.name: *JWT_ES384,
	JWT_ES512.name: *JWT_ES512,

	JWT_EdDSA.name: *JWT_EdDSA,
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
