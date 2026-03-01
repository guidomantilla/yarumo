package passwords

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	cutils "github.com/guidomantilla/yarumo/common/utils"
)

// Decoded structs for multi-value returns.
type argon2Decoded struct {
	prefix     string
	version    int
	iterations int
	memory     int
	threads    int
	salt       []byte
	key        []byte
}

type pbkdf2Decoded struct {
	prefix     string
	iterations int
	salt       []byte
	key        []byte
}

type scryptDecoded struct {
	prefix string
	n      int
	r      int
	p      int
	salt   []byte
	key    []byte
}

func generateSalt(saltSize int) ([]byte, error) {

	unEncodedSalt := make([]byte, saltSize)
	_, err := io.ReadFull(rand.Reader, unEncodedSalt)
	if err != nil {
		return nil, cerrs.Wrap(ErrSaltGenerationFailed, err)
	}

	length := base64.RawStdEncoding.EncodedLen(len(unEncodedSalt))
	encodedSalt := make([]byte, length)
	base64.RawStdEncoding.Encode(encodedSalt, unEncodedSalt)

	return encodedSalt, nil
}

func encode(method *Method, rawPassword string) (string, error) {

	if cutils.Empty(rawPassword) {
		return "", ErrRawPasswordEmpty
	}

	if method.argon2Params != nil {
		return argon2Encode(method, rawPassword)
	}
	if method.bcryptParams != nil {
		return bcryptEncode(method, rawPassword)
	}
	if method.pbkdf2Params != nil {
		return pbkdf2Encode(method, rawPassword)
	}
	if method.scryptParams != nil {
		return scryptEncode(method, rawPassword)
	}

	return "", ErrMethodConfigMissing
}

func verify(method *Method, encodedPassword string, rawPassword string) (bool, error) {

	if cutils.Empty(rawPassword) {
		return false, ErrRawPasswordEmpty
	}

	if cutils.Empty(encodedPassword) {
		return false, ErrEncodedPasswordEmpty
	}

	if method.argon2Params != nil {
		return argon2Verify(method, encodedPassword, rawPassword)
	}
	if method.bcryptParams != nil {
		return bcryptVerify(method, encodedPassword, rawPassword)
	}
	if method.pbkdf2Params != nil {
		return pbkdf2Verify(method, encodedPassword, rawPassword)
	}
	if method.scryptParams != nil {
		return scryptVerify(method, encodedPassword, rawPassword)
	}

	return false, ErrMethodConfigMissing
}

func upgradeNeeded(method *Method, encodedPassword string) (bool, error) {

	if cutils.Empty(encodedPassword) {
		return false, ErrEncodedPasswordEmpty
	}

	if method.argon2Params != nil {
		return argon2UpgradeNeeded(method, encodedPassword)
	}
	if method.bcryptParams != nil {
		return bcryptUpgradeNeeded(method, encodedPassword)
	}
	if method.pbkdf2Params != nil {
		return pbkdf2UpgradeNeeded(method, encodedPassword)
	}
	if method.scryptParams != nil {
		return scryptUpgradeNeeded(method, encodedPassword)
	}

	return false, ErrMethodConfigMissing
}

// --- Argon2 ---

func argon2Encode(method *Method, rawPassword string) (string, error) {
	params := method.argon2Params

	salt, err := generateSalt(params.saltLength)
	if err != nil {
		return "", err
	}

	key := argon2.IDKey([]byte(rawPassword), salt, uint32(params.iterations), uint32(params.memory), uint8(params.threads), uint32(params.keyLength)) //nolint:gosec

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)

	encoded := fmt.Sprintf("%s$%d$%d$%d$%d$%s$%s", method.prefix, argon2.Version, params.iterations, params.memory, params.threads, b64Salt, b64Key)
	return encoded, nil
}

func argon2Decode(encodedPassword string) (*argon2Decoded, error) {

	values := strings.Split(encodedPassword, "$")
	if len(values) != 7 {
		return nil, ErrEncodedPasswordFormat
	}

	prefix := values[0]

	version, err := strconv.Atoi(values[1])
	if err != nil {
		return nil, ErrEncodedPasswordFormat
	}

	iterations, err := strconv.Atoi(values[2])
	if err != nil {
		return nil, ErrEncodedPasswordFormat
	}

	memory, err := strconv.Atoi(values[3])
	if err != nil {
		return nil, ErrEncodedPasswordFormat
	}

	threads, err := strconv.Atoi(values[4])
	if err != nil {
		return nil, ErrEncodedPasswordFormat
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(values[5])
	if err != nil {
		return nil, ErrEncodedPasswordFormat
	}

	key, err := base64.RawStdEncoding.Strict().DecodeString(values[6])
	if err != nil {
		return nil, ErrEncodedPasswordFormat
	}

	return &argon2Decoded{
		prefix:     prefix,
		version:    version,
		iterations: iterations,
		memory:     memory,
		threads:    threads,
		salt:       salt,
		key:        key,
	}, nil
}

func argon2Verify(method *Method, encodedPassword string, rawPassword string) (bool, error) {

	if !strings.HasPrefix(encodedPassword, method.prefix) {
		return false, ErrEncodedPasswordFormat
	}

	decoded, err := argon2Decode(encodedPassword)
	if err != nil {
		return false, err
	}

	newKey := argon2.IDKey([]byte(rawPassword), decoded.salt, uint32(decoded.iterations), uint32(decoded.memory), uint8(decoded.threads), uint32(len(decoded.key))) //nolint:gosec

	return subtle.ConstantTimeCompare(decoded.key, newKey) == 1, nil
}

func argon2UpgradeNeeded(method *Method, encodedPassword string) (bool, error) {

	if !strings.HasPrefix(encodedPassword, method.prefix) {
		return false, ErrEncodedPasswordFormat
	}

	decoded, err := argon2Decode(encodedPassword)
	if err != nil {
		return false, err
	}

	params := method.argon2Params

	if int(argon2.Version) > decoded.version {
		return true, nil
	}
	if params.iterations > decoded.iterations {
		return true, nil
	}
	if params.memory > decoded.memory {
		return true, nil
	}
	if params.threads > decoded.threads {
		return true, nil
	}
	if params.saltLength > len(decoded.salt) {
		return true, nil
	}
	if params.keyLength > len(decoded.key) {
		return true, nil
	}

	return false, nil
}

// --- Bcrypt ---

func bcryptEncode(method *Method, rawPassword string) (string, error) {
	params := method.bcryptParams

	if params.cost < bcrypt.MinCost || params.cost > bcrypt.MaxCost {
		return "", ErrBcryptCostNotAllowed
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), params.cost)
	if err != nil {
		return "", err
	}

	return method.prefix + string(bytes), nil
}

func bcryptVerify(method *Method, encodedPassword string, rawPassword string) (bool, error) {

	if !strings.HasPrefix(encodedPassword, method.prefix) {
		return false, ErrEncodedPasswordFormat
	}

	hash := strings.TrimPrefix(encodedPassword, method.prefix)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(rawPassword))
	if err != nil {
		return false, nil //nolint:nilerr
	}

	return true, nil
}

func bcryptUpgradeNeeded(method *Method, encodedPassword string) (bool, error) {

	if !strings.HasPrefix(encodedPassword, method.prefix) {
		return false, ErrEncodedPasswordFormat
	}

	hash := strings.TrimPrefix(encodedPassword, method.prefix)
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return false, cerrs.Wrap(ErrEncodedPasswordFormat, err)
	}

	return cost < method.bcryptParams.cost, nil
}

// --- Pbkdf2 ---

func pbkdf2Encode(method *Method, rawPassword string) (string, error) {
	params := method.pbkdf2Params

	salt, err := generateSalt(params.saltLength)
	if err != nil {
		return "", err
	}

	bytes := pbkdf2.Key([]byte(rawPassword), salt, params.iterations, params.keyLength, params.hashFunc)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(bytes)

	encoded := fmt.Sprintf("%s$%d$%s$%s", method.prefix, params.iterations, b64Salt, b64Key)
	return encoded, nil
}

func pbkdf2Decode(encodedPassword string) (*pbkdf2Decoded, error) {

	values := strings.Split(encodedPassword, "$")
	if len(values) != 4 {
		return nil, ErrEncodedPasswordFormat
	}

	prefix := values[0]

	iterations, err := strconv.Atoi(values[1])
	if err != nil {
		return nil, ErrEncodedPasswordFormat
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(values[2])
	if err != nil {
		return nil, ErrEncodedPasswordFormat
	}

	key, err := base64.RawStdEncoding.Strict().DecodeString(values[3])
	if err != nil {
		return nil, ErrEncodedPasswordFormat
	}

	return &pbkdf2Decoded{
		prefix:     prefix,
		iterations: iterations,
		salt:       salt,
		key:        key,
	}, nil
}

func pbkdf2Verify(method *Method, encodedPassword string, rawPassword string) (bool, error) {

	if !strings.HasPrefix(encodedPassword, method.prefix) {
		return false, ErrEncodedPasswordFormat
	}

	decoded, err := pbkdf2Decode(encodedPassword)
	if err != nil {
		return false, err
	}

	newKey := pbkdf2.Key([]byte(rawPassword), decoded.salt, decoded.iterations, len(decoded.key), method.pbkdf2Params.hashFunc)

	return subtle.ConstantTimeCompare(decoded.key, newKey) == 1, nil
}

func pbkdf2UpgradeNeeded(method *Method, encodedPassword string) (bool, error) {

	if !strings.HasPrefix(encodedPassword, method.prefix) {
		return false, ErrEncodedPasswordFormat
	}

	decoded, err := pbkdf2Decode(encodedPassword)
	if err != nil {
		return false, err
	}

	params := method.pbkdf2Params

	if params.iterations > decoded.iterations {
		return true, nil
	}
	if params.saltLength > len(decoded.salt) {
		return true, nil
	}
	if params.keyLength > len(decoded.key) {
		return true, nil
	}

	return false, nil
}

// --- Scrypt ---

func scryptEncode(method *Method, rawPassword string) (string, error) {
	params := method.scryptParams

	salt, err := generateSalt(params.saltLength)
	if err != nil {
		return "", err
	}

	bytes, err := scrypt.Key([]byte(rawPassword), salt, params.n, params.r, params.p, params.keyLength)
	if err != nil {
		return "", err
	}

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(bytes)

	encoded := fmt.Sprintf("%s$%d$%d$%d$%s$%s", method.prefix, params.n, params.r, params.p, b64Salt, b64Key)
	return encoded, nil
}

func scryptDecode(encodedPassword string) (*scryptDecoded, error) {

	values := strings.Split(encodedPassword, "$")
	if len(values) != 6 {
		return nil, ErrEncodedPasswordFormat
	}

	prefix := values[0]

	n, err := strconv.Atoi(values[1])
	if err != nil {
		return nil, ErrEncodedPasswordFormat
	}

	r, err := strconv.Atoi(values[2])
	if err != nil {
		return nil, ErrEncodedPasswordFormat
	}

	p, err := strconv.Atoi(values[3])
	if err != nil {
		return nil, ErrEncodedPasswordFormat
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(values[4])
	if err != nil {
		return nil, ErrEncodedPasswordFormat
	}

	key, err := base64.RawStdEncoding.Strict().DecodeString(values[5])
	if err != nil {
		return nil, ErrEncodedPasswordFormat
	}

	return &scryptDecoded{
		prefix: prefix,
		n:      n,
		r:      r,
		p:      p,
		salt:   salt,
		key:    key,
	}, nil
}

func scryptVerify(method *Method, encodedPassword string, rawPassword string) (bool, error) {

	if !strings.HasPrefix(encodedPassword, method.prefix) {
		return false, ErrEncodedPasswordFormat
	}

	decoded, err := scryptDecode(encodedPassword)
	if err != nil {
		return false, err
	}

	newKey, err := scrypt.Key([]byte(rawPassword), decoded.salt, decoded.n, decoded.r, decoded.p, len(decoded.key))
	if err != nil {
		return false, err
	}

	return subtle.ConstantTimeCompare(decoded.key, newKey) == 1, nil
}

func scryptUpgradeNeeded(method *Method, encodedPassword string) (bool, error) {

	if !strings.HasPrefix(encodedPassword, method.prefix) {
		return false, ErrEncodedPasswordFormat
	}

	decoded, err := scryptDecode(encodedPassword)
	if err != nil {
		return false, err
	}

	params := method.scryptParams

	if params.n > decoded.n {
		return true, nil
	}
	if params.r > decoded.r {
		return true, nil
	}
	if params.p > decoded.p {
		return true, nil
	}
	if params.saltLength > len(decoded.salt) {
		return true, nil
	}
	if params.keyLength > len(decoded.key) {
		return true, nil
	}

	return false, nil
}
