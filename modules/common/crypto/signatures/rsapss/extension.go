package rsapss

import (
	"sync"
)

var methods = map[string]Method{
	RSASSA_PSS_using_SHA256.name: *RSASSA_PSS_using_SHA256,
	RSASSA_PSS_using_SHA512.name: *RSASSA_PSS_using_SHA512,
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
