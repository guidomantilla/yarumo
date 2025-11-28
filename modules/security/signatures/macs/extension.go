package macs

import (
	"sync"
)

var algorithms = map[Name]Algorithm{
	HS_256:   {Name: HS_256, Alias: "HS_256", Fn: HMAC_SHA256, KeySize: 32},
	HS3_256:  {Name: HS3_256, Alias: "HS3_256", Fn: HMAC_SHA3_256, KeySize: 32},
	MB2b_256: {Name: MB2b_256, Alias: "MB2b_256", Fn: BLAKE2b_256_MAC, KeySize: 32},
	HS_512:   {Name: HS_512, Alias: "HS_512", Fn: HMAC_SHA512, KeySize: 64},
	HS3_512:  {Name: HS3_512, Alias: "HS3_512", Fn: HMAC_SHA3_512, KeySize: 64},
	MB2b_512: {Name: MB2b_512, Alias: "MB2b_512", Fn: BLAKE2b_512_MAC, KeySize: 64},
}

var lock = new(sync.RWMutex)

func Register(algorithm Algorithm) {
	lock.Lock()
	defer lock.Unlock()

	algorithms[algorithm.Name] = algorithm
}

func Get(name Name) (*Algorithm, error) {
	lock.Lock()
	defer lock.Unlock()

	alg, ok := algorithms[name]
	if !ok {
		return nil, ErrAlgorithmNotSupported
	}

	return &alg, nil
}

func Supported() []Algorithm {
	lock.Lock()
	defer lock.Unlock()

	var list []Algorithm
	for _, alg := range algorithms {
		list = append(list, alg)
	}
	return list
}
