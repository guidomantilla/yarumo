package passwords

import (
	"fmt"

	"github.com/guidomantilla/yarumo/pkg/common/assert"
)

type manager struct {
	encoder   Encoder
	generator Generator
}

func NewManager(encoder Encoder, generator Generator) Manager {
	assert.NotNil(encoder, fmt.Sprintf("%s - error creating: encoder is nil", "rest-client"))
	assert.NotNil(generator, fmt.Sprintf("%s - error creating: generator is nil", "rest-client"))
	return &manager{
		encoder:   encoder,
		generator: generator,
	}
}

func (manager *manager) Encode(rawPassword string) (*string, error) {

	password, err := manager.encoder.Encode(rawPassword)
	if err != nil {
		return nil, ErrPasswordEncodingFailed(err)
	}

	return password, nil
}

func (manager *manager) Matches(encodedPassword string, rawPassword string) (*bool, error) {

	ok, err := manager.encoder.Matches(encodedPassword, rawPassword)
	if err != nil {
		return nil, ErrPasswordMatchingFailed(err)
	}

	return ok, nil
}

func (manager *manager) UpgradeEncoding(encodedPassword string) (*bool, error) {

	ok, err := manager.encoder.UpgradeEncoding(encodedPassword)
	if err != nil {
		return nil, ErrPasswordUpgradeEncodingValidationFailed(err)
	}

	return ok, nil
}

func (manager *manager) Generate() string {
	return manager.generator.Generate()
}

func (manager *manager) Validate(rawPassword string) error {

	err := manager.generator.Validate(rawPassword)
	if err != nil {
		return ErrPasswordValidationFailed(err)
	}

	return nil
}
