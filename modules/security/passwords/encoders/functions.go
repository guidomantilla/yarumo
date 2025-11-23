package encoders

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

var (
	_ GenerateSaltFn = GenerateSalt
	_ Pbkdf2EncodeFn = Pbkdf2Encode
	_ Pbkdf2DecodeFn = Pbkdf2Decode
	_ ScryptEncodeFn = ScryptEncode
	_ ScryptDecodeFn = ScryptDecode
	_ Argon2EncodeFn = Argon2Encode
	_ Argon2DecodeFn = Argon2Decode
)

// Types

type HashFunc func() hash.Hash

type GenerateSaltFn func(saltSize int) ([]byte, error)

type Pbkdf2EncodeFn func(rawPassword string, salt []byte, iterations int, keyLength int, fn HashFunc) (*string, error)

type Pbkdf2DecodeFn func(encodedPassword string) (*string, *int, []byte, []byte, error)

type ScryptEncodeFn func(rawPassword string, salt []byte, N int, r int, p int, keyLen int) (*string, error)

type ScryptDecodeFn func(encodedPassword string) (*string, *int, *int, *int, []byte, []byte, error)

type Argon2EncodeFn func(rawPassword string, salt []byte, iterations int, memory int, threads int, keyLen int) (*string, error)

type Argon2DecodeFn func(encodedPassword string) (*string, *int, *int, *int, *int, []byte, []byte, error)

//  Defaults

func GenerateSalt(saltSize int) ([]byte, error) {

	var err error
	unEncodedSalt := make([]byte, saltSize)
	if _, err = io.ReadFull(rand.Reader, unEncodedSalt); err != nil {
		return nil, err
	}

	length := base64.StdEncoding.EncodedLen(len(unEncodedSalt))
	encodedSalt := make([]byte, length)
	base64.StdEncoding.Encode(encodedSalt, unEncodedSalt)

	return encodedSalt, nil
}

func Pbkdf2Encode(rawPassword string, salt []byte, iterations int, keyLength int, fn HashFunc) (*string, error) {

	if rawPassword == "" {
		return nil, ErrRawPasswordIsEmpty
	}

	if salt == nil {
		return nil, ErrSaltIsNil
	}

	if len(salt) == 0 {
		return nil, ErrSaltIsEmpty
	}

	if fn == nil {
		return nil, ErrHashFuncIsNil
	}

	bytes := pbkdf2.Key([]byte(rawPassword), salt, iterations, keyLength, fn)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(bytes)

	encodedPassword := fmt.Sprintf("$%d$%s$%s", iterations, b64Salt, b64Key)
	return &encodedPassword, nil
}

func Pbkdf2Decode(encodedPassword string) (*string, *int, []byte, []byte, error) {

	if encodedPassword == "" {
		return nil, nil, nil, nil, ErrEncodedPasswordIsEmpty
	}

	values := strings.Split(encodedPassword, "$")
	if len(values) != 4 {
		return nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	var err error
	var prefix string
	if _, err = fmt.Sscanf(values[0], "%s", &prefix); err != nil {
		return nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	var iterations int
	if _, err = fmt.Sscanf(values[1], "%d", &iterations); err != nil {
		return nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	var salt []byte
	salt, err = base64.RawStdEncoding.Strict().DecodeString(values[2])
	if err != nil {
		return nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	var key []byte
	key, err = base64.RawStdEncoding.Strict().DecodeString(values[3])
	if err != nil {
		return nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	return &prefix, &iterations, salt, key, nil

}

func ScryptEncode(rawPassword string, salt []byte, N int, r int, p int, keyLen int) (*string, error) {

	if rawPassword == "" {
		return nil, ErrRawPasswordIsEmpty
	}

	if salt == nil {
		return nil, ErrSaltIsNil
	}

	if len(salt) == 0 {
		return nil, ErrSaltIsEmpty
	}

	var err error
	var bytes []byte
	if bytes, err = scrypt.Key([]byte(rawPassword), salt, N, r, p, keyLen); err != nil {
		return nil, err
	}

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(bytes)

	encodedPassword := fmt.Sprintf("$%d$%d$%d$%s$%s", N, r, p, b64Salt, b64Key)
	return &encodedPassword, nil
}

func ScryptDecode(encodedPassword string) (*string, *int, *int, *int, []byte, []byte, error) {

	if encodedPassword == "" {
		return nil, nil, nil, nil, nil, nil, ErrEncodedPasswordIsEmpty
	}

	values := strings.Split(encodedPassword, "$")
	if len(values) != 6 {
		return nil, nil, nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	var err error
	var prefix string
	if _, err = fmt.Sscanf(values[0], "%s", &prefix); err != nil {
		return nil, nil, nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	var N int
	if _, err = fmt.Sscanf(values[1], "%d", &N); err != nil {
		return nil, nil, nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	var r int
	if _, err = fmt.Sscanf(values[2], "%d", &r); err != nil {
		return nil, nil, nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	var p int
	if _, err = fmt.Sscanf(values[3], "%d", &p); err != nil {
		return nil, nil, nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	var salt []byte
	salt, err = base64.RawStdEncoding.Strict().DecodeString(values[4])
	if err != nil {
		return nil, nil, nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	var key []byte
	key, err = base64.RawStdEncoding.Strict().DecodeString(values[5])
	if err != nil {
		return nil, nil, nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	return &prefix, &N, &r, &p, salt, key, nil

}

func Argon2Encode(rawPassword string, salt []byte, iterations int, memory int, threads int, keyLen int) (*string, error) {

	if rawPassword == "" {
		return nil, ErrRawPasswordIsEmpty
	}

	if salt == nil {
		return nil, ErrSaltIsNil
	}

	if len(salt) == 0 {
		return nil, ErrSaltIsEmpty
	}

	key := argon2.IDKey([]byte(rawPassword), salt, uint32(iterations), uint32(memory), uint8(threads), uint32(keyLen)) //nolint:gosec

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)

	encodedPassword := fmt.Sprintf("$%d$%d$%d$%d$%s$%s", argon2.Version, iterations, memory, threads, b64Salt, b64Key)
	return &encodedPassword, nil
}

func Argon2Decode(encodedPassword string) (*string, *int, *int, *int, *int, []byte, []byte, error) {

	if encodedPassword == "" {
		return nil, nil, nil, nil, nil, nil, nil, ErrEncodedPasswordIsEmpty
	}

	values := strings.Split(encodedPassword, "$")
	if len(values) != 7 {
		return nil, nil, nil, nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	var err error
	var prefix string
	if _, err = fmt.Sscanf(values[0], "%s", &prefix); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	var version int
	if _, err = fmt.Sscanf(values[1], "%d", &version); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	var iterations int
	if _, err = fmt.Sscanf(values[2], "%d", &iterations); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	var memory int
	if _, err = fmt.Sscanf(values[3], "%d", &memory); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	var threads int
	if _, err = fmt.Sscanf(values[4], "%d", &threads); err != nil {
		return nil, nil, nil, nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	var salt []byte
	salt, err = base64.RawStdEncoding.Strict().DecodeString(values[5])
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	var key []byte
	key, err = base64.RawStdEncoding.Strict().DecodeString(values[6])
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, ErrEncodedPasswordNotAllowed
	}

	return &prefix, &version, &iterations, &memory, &threads, salt, key, nil

}
