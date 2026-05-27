package uids

import cuids "github.com/guidomantilla/yarumo/core/common/uids"

// Default UID generators preconfigured with the canonical algorithms
// shipped by this package. Consumers can use them directly or look them
// up by name through the common/uids registry; init() registers them on
// import.
var (
	UuidV4 = cuids.NewUID("UUIDv4", UUIDv4)
	NanoID = cuids.NewUID("NanoID", NANOID)
	Cuid2  = cuids.NewUID("CUID2", CUID2)
	UuidV7 = cuids.NewUID("UUIDv7", UUIDv7)
	Ulid   = cuids.NewUID("ULID", ULID)
	XId    = cuids.NewUID("XID", XID)
)

// init registers the canonical UID generators with the common/uids
// registry so consumers can resolve them via cuids.Lookup by name.
func init() {
	cuids.Register(UuidV4)
	cuids.Register(NanoID)
	cuids.Register(Cuid2)
	cuids.Register(UuidV7)
	cuids.Register(Ulid)
	cuids.Register(XId)
}
