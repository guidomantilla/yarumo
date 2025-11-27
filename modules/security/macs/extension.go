package macs

import (
	"reflect"
	"runtime"
	"sync"
)

var generators = map[string]MacFn{}
var lock = new(sync.RWMutex)

func Register(macFn MacFn) {
	lock.Lock()
	defer lock.Unlock()

	generators[runtime.FuncForPC(reflect.ValueOf(macFn).Pointer()).Name()] = macFn
}

func Get(name string) MacFn {
	lock.Lock()
	defer lock.Unlock()

	generator, ok := generators[name]
	if ok {
		return generator
	}

	return nil
}

func List() []string {
	lock.Lock()
	defer lock.Unlock()

	var names []string
	for alg := range generators {
		names = append(names, alg)
	}
	return names
}

//

func init() {
	Register(HMAC_SHA256)
	Register(HMAC_SHA3_256)
	Register(BLAKE2b_256_MAC)
	Register(HMAC_SHA512)
	Register(HMAC_SHA3_512)
	Register(BLAKE2b_512_MAC)
}
