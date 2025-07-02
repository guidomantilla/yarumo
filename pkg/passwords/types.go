package passwords

var (
	_ GenerateSaltFn = GenerateSalt
	_ Pbkdf2EncodeFn = Pbkdf2Encode
	_ Pbkdf2DecodeFn = Pbkdf2Decode
	_ ScryptEncodeFn = ScryptEncode
	_ ScryptDecodeFn = ScryptDecode
	_ Argon2EncodeFn = Argon2Encode
	_ Argon2DecodeFn = Argon2Decode
)

type GenerateSaltFn func(saltSize int) ([]byte, error)

type Pbkdf2EncodeFn func(rawPassword string, salt []byte, iterations int, keyLength int, fn HashFunc) (*string, error)

type Pbkdf2DecodeFn func(encodedPassword string) (*string, *int, []byte, []byte, error)

type ScryptEncodeFn func(rawPassword string, salt []byte, N int, r int, p int, keyLen int) (*string, error)

type ScryptDecodeFn func(encodedPassword string) (*string, *int, *int, *int, []byte, []byte, error)

type Argon2EncodeFn func(rawPassword string, salt []byte, iterations int, memory int, threads int, keyLen int) (*string, error)

type Argon2DecodeFn func(encodedPassword string) (*string, *int, *int, *int, *int, []byte, []byte, error)

//

var (
	_ Encoder   = (*argon2Encoder)(nil)
	_ Encoder   = (*bcryptEncoder)(nil)
	_ Encoder   = (*pbkdf2Encoder)(nil)
	_ Encoder   = (*scryptEncoder)(nil)
	_ Encoder   = (*manager)(nil)
	_ Generator = (*generator)(nil)
	_ Generator = (*manager)(nil)
	_ Manager   = (*manager)(nil)
)

const (
	Argon2PrefixKey = "{argon2}"
	BcryptPrefixKey = "{bcrypt}"
	Pbkdf2PrefixKey = "{pbkdf2}"
	ScryptPrefixKey = "{scrypt}"
)

type Encoder interface {
	Encode(rawPassword string) (*string, error)
	Matches(encodedPassword string, rawPassword string) (*bool, error)
	UpgradeEncoding(encodedPassword string) (*bool, error)
}

type Generator interface {
	Generate() string
	Validate(rawPassword string) error
}

//

type Manager interface {
	Encoder
	Generator
}
