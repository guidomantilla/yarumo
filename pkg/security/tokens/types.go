package tokens

var (
	_ Generator = (*jwtGenerator)(nil)
)

type Generator interface {
	Generate(subject string, principal Principal) (*string, error)
	Validate(tokenString string) (Principal, error)
}

type Principal map[string]any
