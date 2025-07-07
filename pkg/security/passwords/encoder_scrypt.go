package passwords

import (
	"strings"
)

type scryptEncoder struct {
	n          int
	r          int
	p          int
	saltLength int
	keyLength  int
}

func NewScryptEncoder(opts ...ScryptEncoderOption) Encoder {
	options := NewScryptEncoderOptions(opts...)
	return &scryptEncoder{
		n:          options.n,
		r:          options.r,
		p:          options.p,
		saltLength: options.saltLength,
		keyLength:  options.keyLength,
	}

}

func (encoder *scryptEncoder) Encode(rawPassword string) (*string, error) {

	if rawPassword == "" {
		return nil, ErrRawPasswordIsEmpty
	}

	salt, err := GenerateSalt(encoder.saltLength)
	if err != nil {
		return nil, err
	}

	value, err := ScryptEncode(rawPassword, salt, encoder.n, encoder.r, encoder.p, encoder.keyLength)
	if err != nil {
		return nil, err
	}

	encodedPassword := *value
	encodedPassword = ScryptPrefixKey + encodedPassword
	return &encodedPassword, nil
}

func (encoder *scryptEncoder) Matches(encodedPassword string, rawPassword string) (*bool, error) {

	if rawPassword == "" {
		return nil, ErrRawPasswordIsEmpty
	}

	if encodedPassword == "" {
		return nil, ErrEncodedPasswordIsEmpty
	}

	if !strings.HasPrefix(encodedPassword, ScryptPrefixKey) {
		return nil, ErrEncodedPasswordNotAllowed
	}

	_, N, r, p, salt, key, err := ScryptDecode(encodedPassword)
	if err != nil {
		return nil, err
	}

	newEncodedPassword, err := ScryptEncode(rawPassword, salt, *N, *r, *p, len(key))
	if err != nil {
		return nil, err
	}

	encodedPassword = strings.Replace(encodedPassword, ScryptPrefixKey, "", 1)
	matched := encodedPassword == *(newEncodedPassword)
	return &matched, nil
}

func (encoder *scryptEncoder) UpgradeEncoding(encodedPassword string) (*bool, error) {

	if encodedPassword == "" {
		return nil, ErrEncodedPasswordIsEmpty
	}

	if !strings.HasPrefix(encodedPassword, ScryptPrefixKey) {
		return nil, ErrEncodedPasswordNotAllowed
	}

	_, N, r, p, salt, key, err := ScryptDecode(encodedPassword)
	if err != nil {
		return nil, err
	}

	upgradeNeeded := true
	if encoder.n > *(N) {
		return &upgradeNeeded, nil
	}

	if encoder.r > *(r) {
		return &upgradeNeeded, nil
	}

	if encoder.p > *(p) {
		return &upgradeNeeded, nil
	}

	if encoder.saltLength > len(salt) {
		return &upgradeNeeded, nil
	}

	if encoder.keyLength > len(key) {
		return &upgradeNeeded, nil
	}

	upgradeNeeded = false
	return &upgradeNeeded, nil
}
