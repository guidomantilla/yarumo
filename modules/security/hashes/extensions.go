package hashes

import (
	"crypto"
	"sync"
)

var methods = map[string]Method{
	SHA256.name:      *SHA256,
	SHA512.name:      *SHA512,
	SHA3_256.name:    *SHA3_256,
	SHA3_512.name:    *SHA3_512,
	BLAKE2b_256.name: *BLAKE2b_256,
	BLAKE2b_512.name: *BLAKE2b_512,
}

var lock = new(sync.RWMutex)

func Register(method Method) {
	lock.Lock()
	defer lock.Unlock()

	methods[method.name] = method
	crypto.RegisterHash(method.kind, method.fn)
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
