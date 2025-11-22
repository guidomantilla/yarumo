package tokens

var (
	_ GenerateFn = JwtGenerate
	_ ValidateFn = JwtValidate

	_ GenerateFn = OpaqueGenerate
	_ ValidateFn = OpaqueValidate

	_ GenerateFn = JwtGenerator.Generate
	_ ValidateFn = JwtGenerator.Validate

	_ GenerateFn = OpaqueGenerator.Generate
	_ ValidateFn = OpaqueGenerator.Validate
)

// Types

type GenerateFn func(subject string, principal Principal) (*string, error)

type ValidateFn func(tokenString string) (Principal, error)

// Defaults

func JwtGenerate(subject string, principal Principal) (*string, error) {
	return JwtGenerator.Generate(subject, principal)
}

func JwtValidate(tokenString string) (Principal, error) {
	return JwtGenerator.Validate(tokenString)
}

func OpaqueGenerate(subject string, principal Principal) (*string, error) {
	return OpaqueGenerator.Generate(subject, principal)
}

func OpaqueValidate(tokenString string) (Principal, error) {
	return OpaqueGenerator.Validate(tokenString)
}
