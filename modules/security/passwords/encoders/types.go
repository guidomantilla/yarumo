package encoders

const (
	Argon2PrefixKey = "{argon2}"
	BcryptPrefixKey = "{bcrypt}"
	Pbkdf2PrefixKey = "{pbkdf2}"
	ScryptPrefixKey = "{scrypt}"
)

type Encoder interface {
	Encode(rawPassword string) (*string, error)
	Matches(encodedPassword string, rawPassword string) (*bool, error)
	Upgrade(encodedPassword string) (*bool, error)
}
