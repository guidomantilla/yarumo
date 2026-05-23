// Package uids provides pluggable unique identifier generation with support
// for multiple algorithms including UUID, ULID, NanoID, CUID2, and XID.
package uids

var (
	_ UID   = (*uid)(nil)
	_ error = (*Error)(nil)

	_ UIDFn                      = UUIDv4
	_ UIDFn                      = NANOID
	_ UIDFn                      = CUID2
	_ UIDFn                      = UUIDv7
	_ UIDFn                      = ULID
	_ UIDFn                      = XID
	_ IsUIDFn                    = IsUUID
	_ IsUIDFn                    = IsULID
	_ IsUIDFn                    = IsNanoID
	_ IsUIDFn                    = IsCUID2
	_ IsUIDFn                    = IsXID
	_ RegisterFn                 = Register
	_ LookupFn                   = Lookup
	_ SupportedFn                = Supported
	_ ErrAlgorithmNotSupportedFn = ErrAlgorithmNotSupported
)

// UID defines the interface for a named unique identifier generator.
type UID interface {
	// Name returns the algorithm name.
	Name() string
	// Generate generates and returns a new unique identifier, or an error if
	// the underlying entropy source fails.
	Generate() (string, error)
}

// UIDFn is the function type for UID generation functions. Implementations
// return an error when the underlying entropy source (typically crypto/rand)
// fails. Silent fallbacks are not permitted: an empty string with a nil error
// is never acceptable.
type UIDFn func() (string, error)

// IsUIDFn is the function type for UID format validators. Implementations
// report whether the input string matches the canonical format of a
// specific algorithm, without parsing it into a structured value.
type IsUIDFn func(s string) bool

// RegisterFn is the function type for Register.
type RegisterFn func(uid UID)

// LookupFn is the function type for Get.
type LookupFn func(name string) (UID, error)

// SupportedFn is the function type for Supported.
type SupportedFn func() []UID

// ErrAlgorithmNotSupportedFn is the function type for ErrAlgorithmNotSupported.
type ErrAlgorithmNotSupportedFn func(name string) error
