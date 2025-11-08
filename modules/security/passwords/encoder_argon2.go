package passwords

import (
	"strings"

	"golang.org/x/crypto/argon2"
)

type argon2Encoder struct {
	iterations int
	memory     int
	threads    int
	saltLength int
	keyLength  int
}

func NewArgon2Encoder(opts ...Argon2EncoderOption) Encoder {
	options := NewArgon2EncoderOptions(opts...)
	return &argon2Encoder{
		iterations: options.iterations,
		memory:     options.memory,
		threads:    options.threads,
		saltLength: options.saltLength,
		keyLength:  options.keyLength,
	}
}

func (encoder *argon2Encoder) Encode(rawPassword string) (*string, error) {

	if rawPassword == "" {
		return nil, ErrRawPasswordIsEmpty
	}

	salt, err := GenerateSalt(encoder.saltLength)
	if err != nil {
		return nil, err
	}

	value, err := Argon2Encode(rawPassword, salt, encoder.iterations, encoder.memory, encoder.threads, encoder.keyLength)
	if err != nil {
		return nil, err
	}

	encoded := *value
	encoded = Argon2PrefixKey + encoded
	return &encoded, nil
}

func (encoder *argon2Encoder) Matches(encodedPassword string, rawPassword string) (*bool, error) {

	if rawPassword == "" {
		return nil, ErrRawPasswordIsEmpty
	}

	if encodedPassword == "" {
		return nil, ErrEncodedPasswordIsEmpty
	}

	if !strings.HasPrefix(encodedPassword, Argon2PrefixKey) {
		return nil, ErrEncodedPasswordNotAllowed
	}

	_, _, iterations, memory, threads, salt, key, err := Argon2Decode(encodedPassword)
	if err != nil {
		return nil, err
	}

	newEncoded, err := Argon2Encode(rawPassword, salt, *iterations, *memory, *threads, len(key))
	if err != nil {
		return nil, err
	}

	encodedPassword = strings.Replace(encodedPassword, Argon2PrefixKey, "", 1)
	matched := encodedPassword == *(newEncoded)
	return &matched, nil
}

func (encoder *argon2Encoder) UpgradeEncoding(encodedPassword string) (*bool, error) {

	if encodedPassword == "" {
		return nil, ErrRawPasswordIsEmpty
	}

	if !strings.HasPrefix(encodedPassword, Argon2PrefixKey) {
		return nil, ErrEncodedPasswordNotAllowed
	}

	_, version, iterations, memory, threads, salt, key, err := Argon2Decode(encodedPassword)
	if err != nil {
		return nil, err
	}

	upgradeNeeded := true
	if argon2.Version > *(version) {
		return &upgradeNeeded, nil
	}

	if encoder.iterations > *(iterations) {
		return &upgradeNeeded, nil
	}

	if encoder.memory > *(memory) {
		return &upgradeNeeded, nil
	}

	if encoder.threads > *(threads) {
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
