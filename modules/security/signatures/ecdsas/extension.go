package ecdsas

import (
	"sync"
)

var methods = map[string]Method{
	ECDSA_with_SHA256_over_P256.name: *ECDSA_with_SHA256_over_P256,
	ECDSA_with_SHA512_over_P521.name: *ECDSA_with_SHA512_over_P521,
}

var lock = new(sync.RWMutex)

func Register(method Method) {
	lock.Lock()
	defer lock.Unlock()

	methods[method.name] = method
}

func Get(name string) (*Method, error) {
	lock.Lock()
	defer lock.Unlock()

	alg, ok := methods[name]
	if !ok {
		return nil, ErrAlgorithmNotSupported(name)
	}
	return &alg, nil
}

func Supported() []Method {
	lock.Lock()
	defer lock.Unlock()

	var list []Method
	for _, method := range methods {
		list = append(list, method)
	}
	return list
}
