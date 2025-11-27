package tokens

import "sync"

var generators = map[string]Generator{}
var lock = new(sync.RWMutex)

func Register(generator Generator) {
	lock.Lock()
	defer lock.Unlock()

	generators[generator.Name()] = generator
}

func Get(name string) Generator {
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

var (
	DefaultJwtGenerator    = NewJwtGenerator()
	DefaultOpaqueGenerator = NewOpaqueGenerator()
)

func init() {
	Register(DefaultJwtGenerator)
	Register(DefaultOpaqueGenerator)
}
