package tokens

import "sync"

var (
	DefaultJwtGenerator    = NewJwtGenerator()
	DefaultOpaqueGenerator = NewOpaqueGenerator()
)

var algorithms = map[Name]Algorithm{
	DefaultJwtGenerator.Name():    {Name: DefaultJwtGenerator.Name(), Generator: DefaultJwtGenerator},
	DefaultOpaqueGenerator.Name(): {Name: DefaultOpaqueGenerator.Name(), Generator: DefaultOpaqueGenerator},
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
