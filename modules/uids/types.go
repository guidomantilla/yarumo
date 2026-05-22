// Package uids provides a provider-agnostic registry of unique identifier
// generators. Provider implementations live in sub-modules under
// modules/uids/<provider>/ (cuid2, nanoid, uuid, ulid, xid). Each provider
// sub-module exposes preconfigured singletons (uuid.UuidV4, uuid.UuidV7,
// cuid2.Cuid2, etc.) and free functions; consumers use them directly or
// register them explicitly with this registry when they want name-based
// Lookup:
//
//	import (
//	    "github.com/guidomantilla/yarumo/uids"
//	    "github.com/guidomantilla/yarumo/uids/uuid"
//	)
//
//	uids.Register(uuid.UuidV4)
//	uids.Register(uuid.UuidV7)
//
// There is intentionally no init()-based auto-registration: side effects
// on import are not used. This split keeps modules/uids/ free of
// third-party provider dependencies so callers only pay for the providers
// they actually import.
package uids

var (
	_ UID   = (*uid)(nil)
	_ error = (*Error)(nil)

	_ NewUIDFn                   = NewUID
	_ RegisterFn                 = Register
	_ LookupFn                   = Lookup
	_ SupportedFn                = Supported
	_ ErrAlgorithmNotSupportedFn = ErrAlgorithmNotSupported
	_ ErrGenerationFn            = ErrGeneration
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

// NewUIDFn is the function type for NewUID.
type NewUIDFn func(name string, fn UIDFn) UID

// RegisterFn is the function type for Register.
type RegisterFn func(uid UID)

// LookupFn is the function type for Lookup.
type LookupFn func(name string) (UID, error)

// SupportedFn is the function type for Supported.
type SupportedFn func() []UID

// ErrAlgorithmNotSupportedFn is the function type for ErrAlgorithmNotSupported.
type ErrAlgorithmNotSupportedFn func(name string) error

// ErrGenerationFn is the function type for ErrGeneration.
type ErrGenerationFn func(errs ...error) error
