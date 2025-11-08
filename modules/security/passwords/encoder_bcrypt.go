package passwords

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type bcryptEncoder struct {
	cost int
}

func NewBcryptEncoder(opts ...BcryptEncoderOption) Encoder {
	options := NewBcryptEncoderOptions(opts...)
	return &bcryptEncoder{cost: options.cost}
}

func (encoder *bcryptEncoder) Encode(rawPassword string) (*string, error) {

	if rawPassword == "" {
		return nil, ErrRawPasswordIsEmpty
	}

	if encoder.cost < bcrypt.MinCost || encoder.cost > bcrypt.MaxCost {
		return nil, ErrBcryptCostNotAllowed
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), encoder.cost)
	if err != nil {
		return nil, err
	}

	encodedPassword := BcryptPrefixKey + string(bytes)
	return &encodedPassword, nil
}

func (encoder *bcryptEncoder) Matches(encodedPassword string, rawPassword string) (*bool, error) {

	if rawPassword == "" {
		return nil, ErrRawPasswordIsEmpty
	}

	if encodedPassword == "" {
		return nil, ErrEncodedPasswordIsEmpty
	}

	if !strings.HasPrefix(encodedPassword, BcryptPrefixKey) {
		return nil, ErrEncodedPasswordNotAllowed
	}

	matched := true
	encodedPassword = strings.Replace(encodedPassword, BcryptPrefixKey, "", 1)
	err := bcrypt.CompareHashAndPassword([]byte(encodedPassword), []byte(rawPassword))
	if err != nil {
		matched = false
	}

	return &matched, nil
}

func (encoder *bcryptEncoder) UpgradeEncoding(encodedPassword string) (*bool, error) {

	if encodedPassword == "" {
		return nil, ErrEncodedPasswordIsEmpty
	}

	if !strings.HasPrefix(encodedPassword, BcryptPrefixKey) {
		return nil, ErrEncodedPasswordNotAllowed
	}

	encodedPassword = strings.Replace(encodedPassword, BcryptPrefixKey, "", 1)

	cost, _ := bcrypt.Cost([]byte(encodedPassword))
	upgradeNeeded := cost < encoder.cost

	return &upgradeNeeded, nil
}
