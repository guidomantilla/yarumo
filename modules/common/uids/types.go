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
	_ RegisterFn                 = Register
	_ GetFn                      = Get
	_ UseFn                      = Use
	_ GenerateFn                 = Generate
	_ SupportedFn                = Supported
	_ ErrAlgorithmNotSupportedFn = ErrAlgorithmNotSupported
)

// UID defines the interface for a named unique identifier generator.
type UID interface {
	// Name returns the algorithm name.
	Name() string
	// Generate generates and returns a new unique identifier.
	Generate() string
}

// UIDFn is the function type for UID generation functions.
type UIDFn func() string

// RegisterFn is the function type for Register.
type RegisterFn func(uid UID)

// GetFn is the function type for Get.
type GetFn func(name string) (UID, error)

// UseFn is the function type for Use.
type UseFn func(name string) error

// GenerateFn is the function type for Generate.
type GenerateFn func() string

// SupportedFn is the function type for Supported.
type SupportedFn func() []UID

// ErrAlgorithmNotSupportedFn is the function type for ErrAlgorithmNotSupported.
type ErrAlgorithmNotSupportedFn func(name string) error
