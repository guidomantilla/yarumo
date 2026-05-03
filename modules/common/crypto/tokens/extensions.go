package tokens

import "sync"

var methods = map[string]Method{
	JWT_HS256.name: *JWT_HS256,
	JWT_HS384.name: *JWT_HS384,
	JWT_HS512.name: *JWT_HS512,
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
