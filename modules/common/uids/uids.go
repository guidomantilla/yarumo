package uids

import cassert "github.com/guidomantilla/yarumo/common/assert"

// Default UID generators preconfigured with standard algorithms.
var (
	UuidV4 = NewUID("UUIDv4", UUIDv4)
	NanoID = NewUID("NanoID", NANOID)
	Cuid2  = NewUID("CUID2", CUID2)
	UuidV7 = NewUID("UUIDv7", UUIDv7)
	Ulid   = NewUID("ULID", ULID)
	XId    = NewUID("XID", XID)
)

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

// Generate generates and returns a new unique identifier.
func (u *uid) Generate() string {
	cassert.NotNil(u, "uid is nil")
	return u.fn()
}
