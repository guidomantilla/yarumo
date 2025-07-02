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

func (encoder *argon2Encoder) Encode(raw string) (*string, error) {

	salt, err := GenerateSalt(encoder.saltLength)
	if err != nil {
		return nil, err
	}

	value, err := Argon2Encode(raw, salt, encoder.iterations, encoder.memory, encoder.threads, encoder.keyLength)
	if err != nil {
		return nil, err
	}

	encoded := *value
	encoded = Argon2PrefixKey + encoded
	return &encoded, nil
}

func (encoder *argon2Encoder) Matches(encoded string, raw string) (*bool, error) {

	if raw == "" {
		return nil, ErrRawPasswordIsEmpty
	}

	if !strings.HasPrefix(encoded, Argon2PrefixKey) {
		return nil, ErrEncodedPasswordNotAllowed
	}

	_, _, iterations, memory, threads, salt, key, err := Argon2Decode(encoded)
	if err != nil {
		return nil, err
	}

	newEncoded, err := Argon2Encode(raw, salt, *iterations, *memory, *threads, len(key))
	if err != nil {
		return nil, err
	}

	encoded = strings.Replace(encoded, Argon2PrefixKey, "", 1)
	matched := encoded == *(newEncoded)
	return &matched, nil
}

func (encoder *argon2Encoder) UpgradeEncoding(encoded string) (*bool, error) {

	if encoded == "" {
		return nil, ErrRawPasswordIsEmpty
	}

	if !strings.HasPrefix(encoded, Argon2PrefixKey) {
		return nil, ErrEncodedPasswordNotAllowed
	}

	_, version, iterations, memory, threads, salt, key, err := Argon2Decode(encoded)
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
