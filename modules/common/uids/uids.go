package uids

var (
	UuidV4 = NewUID("UUIDv4", UUIDv4)
	NanoID = NewUID("NanoID", NANOID)
	Cuid2  = NewUID("CUID2", CUID2)
	UuidV7 = NewUID("UUIDv7", UUIDv7)
	Ulid   = NewUID("ULID", ULID)
	XId    = NewUID("XID", XID)
)

type UID struct {
	name string
	fn   UIDFn
}

func NewUID(name string, fn UIDFn) *UID {
	return &UID{name: name, fn: fn}
}

func (u *UID) Name() string {
	return u.name
}

func (u *UID) Generate() string {
	return u.fn()
}
