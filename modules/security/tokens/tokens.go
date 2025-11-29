package tokens

import "github.com/guidomantilla/yarumo/common/assert"

var (
	DefaultJwtGenerator    = NewJwtGenerator()
	DefaultOpaqueGenerator = NewOpaqueGenerator()
)

type Algorithm struct {
	name      string
	generator Generator
}

func NewAlgorithm(name string, generator Generator) *Algorithm {
	return &Algorithm{
		name:      name,
		generator: generator,
	}
}

func (a *Algorithm) Name() string {
	assert.NotNil(a, "algorithm is nil")
	return a.name
}

func (a *Algorithm) Generate(subject string, principal Principal) (string, error) {
	assert.NotNil(a, "algorithm is nil")
	return a.generator.Generate(subject, principal)
}

func (a *Algorithm) Validate(tokenString string) (Principal, error) {
	assert.NotNil(a, "algorithm is nil")
	return a.generator.Validate(tokenString)
}
