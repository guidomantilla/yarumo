package aead

import (
	"sync"
)

var methods = map[string]Method{
	AES_128_GCM.name:        *AES_128_GCM,
	AES_256_GCM.name:        *AES_256_GCM,
	CHACHA20_POLY1305.name:  *CHACHA20_POLY1305,
	XCHACHA20_POLY1305.name: *XCHACHA20_POLY1305,
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
