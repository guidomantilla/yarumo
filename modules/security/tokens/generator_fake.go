package tokens

import (
	"github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/common/utils"
)

type FakeGenerator struct {
	GenerateFn GenerateFn
	ValidateFn ValidateFn
}

func (g *FakeGenerator) Generate(subject string, principal Principal) (*string, error) {
	assert.NotNil(g, "generator is nil")
	assert.NotNil(g.GenerateFn, "GenerateFn is nil")

	if utils.Empty(subject) {
		return nil, ErrTokenGeneration(ErrSubjectCannotBeEmpty)
	}
	if utils.Empty(principal) {
		return nil, ErrTokenGeneration(ErrPrincipalCannotBeNil)
	}

	return g.GenerateFn(subject, principal)
}

func (g *FakeGenerator) Validate(tokenString string) (Principal, error) {
	assert.NotNil(g, "generator is nil")
	assert.NotNil(g.ValidateFn, "ValidateFn is nil")

	if utils.Empty(tokenString) {
		return nil, ErrTokenValidation(ErrTokenCannotBeEmpty)
	}

	return g.ValidateFn(tokenString)
}
