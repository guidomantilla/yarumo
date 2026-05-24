// Package uids declares the contract for pluggable unique identifier
// generation: the UID interface, the function-type aliases, a generic
// registry (Register/Lookup/Supported), and a trivial constructor NewUID
// that wraps any UIDFn into a UID value.
//
// This package carries NO concrete generator implementations and NO
// external dependencies — those live in modules/extensions/common/uids/,
// which registers the canonical algorithms (UUIDv4/UUIDv7/ULID/NanoID/
// CUID2/XID) into this package's registry via its package init().
// Consumers that only need the abstract contract or to register custom
// UIDs can import this package alone; consumers that need the catalogue
// of well-known algorithms import the extensions package.
package uids

var (
	_ UID   = (*uid)(nil)
	_ error = (*Error)(nil)

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

// LookupFn is the function type for Lookup.
type LookupFn func(name string) (UID, error)

// SupportedFn is the function type for Supported.
type SupportedFn func() []UID

// ErrAlgorithmNotSupportedFn is the function type for ErrAlgorithmNotSupported.
type ErrAlgorithmNotSupportedFn func(name string) error
