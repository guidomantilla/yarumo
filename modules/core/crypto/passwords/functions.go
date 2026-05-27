package passwords

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"

	crandom "github.com/guidomantilla/yarumo/core/crypto/random"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"
	cutils "github.com/guidomantilla/yarumo/core/common/utils"
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

func encode(method *Method, rawPassword string) (string, error) {
	if method == nil {
		return "", ErrMethodIsNil
	}

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
	if method == nil {
		return false, ErrMethodIsNil
	}

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
	if method == nil {
		return false, ErrMethodIsNil
	}

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

	salt, err := crandom.Bytes(params.saltLength)
	if err != nil {
		return "", cerrs.Wrap(ErrSaltGenerationFailed, err)
	}

	if len(salt) == 0 {
		return "", ErrSaltGenerationFailed
	}

	key := argon2DeriveKey(params.useArgon2i, []byte(rawPassword), salt, params.iterations, params.memory, params.threads, params.keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)

	encoded := fmt.Sprintf("%s$%d$%d$%d$%d$%s$%s", method.prefix, argon2.Version, params.iterations, params.memory, params.threads, b64Salt, b64Key)
	return encoded, nil
}

// argon2DeriveKey dispatches to argon2.Key (argon2i, side-channel resistant)
// or argon2.IDKey (argon2id, OWASP recommended) based on useArgon2i.
func argon2DeriveKey(useArgon2i bool, password, salt []byte, iterations, memory, threads, keyLength int) []byte {
	if useArgon2i {
		return argon2.Key(password, salt, uint32(iterations), uint32(memory), uint8(threads), uint32(keyLength)) //nolint:gosec
	}
	return argon2.IDKey(password, salt, uint32(iterations), uint32(memory), uint8(threads), uint32(keyLength)) //nolint:gosec
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

	if !argon2PrefixMatches(method, encodedPassword) {
		return false, ErrEncodedPasswordFormat
	}

	decoded, err := argon2Decode(encodedPassword)
	if err != nil {
		return false, err
	}

	newKey := argon2DeriveKey(method.argon2Params.useArgon2i, []byte(rawPassword), decoded.salt, decoded.iterations, decoded.memory, decoded.threads, len(decoded.key))

	return subtle.ConstantTimeCompare(decoded.key, newKey) == 1, nil
}

// argon2PrefixMatches returns true if encodedPassword carries the method's
// own prefix, OR — when the method is the Argon2id default (prefix
// {argon2id}) — the legacy {argon2} prefix. This preserves verification of
// stored hashes produced by the pre-YA-0030 code, which always emitted
// {argon2} regardless of the underlying argon2.IDKey call.
func argon2PrefixMatches(method *Method, encodedPassword string) bool {
	if strings.HasPrefix(encodedPassword, method.prefix) {
		return true
	}
	if method.prefix == Argon2idPrefixKey && strings.HasPrefix(encodedPassword, Argon2PrefixKey) {
		return true
	}
	return false
}

func argon2UpgradeNeeded(method *Method, encodedPassword string) (bool, error) {

	if !argon2PrefixMatches(method, encodedPassword) {
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

	salt, err := crandom.Bytes(params.saltLength)
	if err != nil {
		return "", cerrs.Wrap(ErrSaltGenerationFailed, err)
	}

	if len(salt) == 0 {
		return "", ErrSaltGenerationFailed
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

	salt, saltErr := crandom.Bytes(params.saltLength)
	if saltErr != nil {
		return "", cerrs.Wrap(ErrSaltGenerationFailed, saltErr)
	}

	if len(salt) == 0 {
		return "", ErrSaltGenerationFailed
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
