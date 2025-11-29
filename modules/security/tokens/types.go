package tokens

var (
	_ Generator = (*jwtGenerator)(nil)
	_ Generator = (*opaqueGenerator)(nil)
	_ Generator = (*FakeGenerator)(nil)
)

type Generator interface {
	Name() string
	Generate(subject string, principal Principal) (string, error)
	Validate(tokenString string) (Principal, error)
}

type Principal map[string]any
