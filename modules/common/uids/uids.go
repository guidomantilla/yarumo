package uids

import cassert "github.com/guidomantilla/yarumo/common/assert"

// uid implements the UID interface.
type uid struct {
	name string
	fn   UIDFn
}

// NewUID creates a new UID with the given name and generation function.
func NewUID(name string, fn UIDFn) UID {
	cassert.NotEmpty(name, "name is empty")
	cassert.NotNil(fn, "fn is nil")

	return &uid{name: name, fn: fn}
}

// Name returns the algorithm name.
func (u *uid) Name() string {
	cassert.NotNil(u, "uid is nil")
	return u.name
}

// Generate generates and returns a new unique identifier, or an error if the
// underlying entropy source fails.
func (u *uid) Generate() (string, error) {
	cassert.NotNil(u, "uid is nil")
	return u.fn()
}
