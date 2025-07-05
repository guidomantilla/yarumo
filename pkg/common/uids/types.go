package uids

var (
	_ UIDFn = UUIDv4
	_ UIDFn = NANOID
	_ UIDFn = CUID2
	_ UIDFn = UUIDv7
	_ UIDFn = ULID
	_ UIDFn = XID
)

type UIDFn func() string
