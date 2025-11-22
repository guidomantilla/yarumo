package tokens

var (
	_ GenerateFn = JwtGenerate
	_ ValidateFn = JwtValidate

	_ GenerateFn = OpaqueGenerate
	_ ValidateFn = OpaqueValidate

	_ GenerateFn = DefaultJwtGenerator.Generate
	_ ValidateFn = DefaultJwtGenerator.Validate

	_ GenerateFn = DefaultOpaqueGenerator.Generate
	_ ValidateFn = DefaultOpaqueGenerator.Validate
)

// Types

type GenerateFn func(subject string, principal Principal) (*string, error)

type ValidateFn func(tokenString string) (Principal, error)

// Defaults

func JwtGenerate(subject string, principal Principal) (*string, error) {
	return DefaultJwtGenerator.Generate(subject, principal)
}

func JwtValidate(tokenString string) (Principal, error) {
	return DefaultJwtGenerator.Validate(tokenString)
}

func OpaqueGenerate(subject string, principal Principal) (*string, error) {
	return DefaultOpaqueGenerator.Generate(subject, principal)
}

func OpaqueValidate(tokenString string) (Principal, error) {
	return DefaultOpaqueGenerator.Validate(tokenString)
}
