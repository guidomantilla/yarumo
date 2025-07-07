package passwords

import (
	"crypto/sha512"
	"hash"
	"strings"
)

var (
	_ HashFunc = sha512.New
	_ HashFunc = sha512.New384
	_ HashFunc = sha512.New512_224
	_ HashFunc = sha512.New512_256
)

type HashFunc func() hash.Hash

type pbkdf2Encoder struct {
	iterations int
	saltLength int
	keyLength  int
	hashFunc   HashFunc
}

func NewPbkdf2Encoder(opts ...Pbkdf2EncoderOption) Encoder {
	options := NewPbkdf2EncoderOptions(opts...)
	return &pbkdf2Encoder{
		iterations: options.iterations,
		saltLength: options.saltLength,
		keyLength:  options.keyLength,
		hashFunc:   options.hashFunc,
	}
}

func (encoder *pbkdf2Encoder) Encode(rawPassword string) (*string, error) {

	if rawPassword == "" {
		return nil, ErrRawPasswordIsEmpty
	}

	salt, err := GenerateSalt(encoder.saltLength)
	if err != nil {
		return nil, err
	}

	value, err := Pbkdf2Encode(rawPassword, salt, encoder.iterations, encoder.keyLength, encoder.hashFunc)
	if err != nil {
		return nil, err
	}

	encodedPassword := *value
	encodedPassword = Pbkdf2PrefixKey + encodedPassword
	return &encodedPassword, nil
}

func (encoder *pbkdf2Encoder) Matches(encodedPassword string, rawPassword string) (*bool, error) {

	if rawPassword == "" {
		return nil, ErrRawPasswordIsEmpty
	}

	if encodedPassword == "" {
		return nil, ErrEncodedPasswordIsEmpty
	}

	if !strings.HasPrefix(encodedPassword, Pbkdf2PrefixKey) {
		return nil, ErrEncodedPasswordNotAllowed
	}

	_, iterations, salt, key, err := Pbkdf2Decode(encodedPassword)
	if err != nil {
		return nil, err
	}

	newEncodedPassword, err := Pbkdf2Encode(rawPassword, salt, *iterations, len(key), encoder.hashFunc)
	if err != nil {
		return nil, err
	}

	encodedPassword = strings.Replace(encodedPassword, Pbkdf2PrefixKey, "", 1)
	matched := encodedPassword == *(newEncodedPassword)
	return &matched, nil
}

func (encoder *pbkdf2Encoder) UpgradeEncoding(encodedPassword string) (*bool, error) {

	if encodedPassword == "" {
		return nil, ErrEncodedPasswordIsEmpty
	}

	if !strings.HasPrefix(encodedPassword, Pbkdf2PrefixKey) {
		return nil, ErrEncodedPasswordNotAllowed
	}

	_, iterations, salt, key, err := Pbkdf2Decode(encodedPassword)
	if err != nil {
		return nil, err
	}

	upgradeNeeded := true
	if encoder.iterations > *(iterations) {
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
