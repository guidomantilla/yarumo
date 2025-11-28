package tokens

type Name string

var (
	_ Generator = (*jwtGenerator)(nil)
	_ Generator = (*opaqueGenerator)(nil)
	_ Generator = (*FakeGenerator)(nil)
)

type Generator interface {
	Name() Name
	Generate(subject string, principal Principal) (string, error)
	Validate(tokenString string) (Principal, error)
}

type Principal map[string]any

type Algorithm struct {
	Name      Name      `json:"name"`
	Generator Generator `json:"-"`
}
