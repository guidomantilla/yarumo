package hashes

import (
	"crypto"
	"sync"
)

var algorithms = map[string]Algorithm{
	SHA256.name:      *SHA256,
	SHA512.name:      *SHA512,
	SHA3_256.name:    *SHA3_256,
	SHA3_512.name:    *SHA3_512,
	BLAKE2b_256.name: *BLAKE2b_256,
	BLAKE2b_512.name: *BLAKE2b_512,
}

var lock = new(sync.RWMutex)

func Register(algorithm Algorithm) {
	lock.Lock()
	defer lock.Unlock()

	algorithms[algorithm.name] = algorithm
	crypto.RegisterHash(algorithm.kind, algorithm.fn)
}

func Get(name string) (*Algorithm, error) {
	lock.Lock()
	defer lock.Unlock()

	alg, ok := algorithms[name]
	if !ok {
		return nil, ErrAlgorithmNotSupported(name)
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
