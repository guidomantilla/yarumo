package passwords

import (
	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// DelegatingEncoder is a Spring-Security-style password encoder that encodes
// new passwords with a configurable primary Method but routes Verify and
// UpgradeNeeded calls to whichever method matches the encoded prefix via the
// package registry.
//
// This enables gradual algorithm migration: hashes produced by a legacy
// algorithm (e.g. Bcrypt) still verify correctly, and UpgradeNeeded signals
// when the caller should re-encode a successfully verified password under the
// new primary algorithm — the canonical "login-time upgrade" pattern.
type DelegatingEncoder struct {
	primary *Method
}

// NewDelegatingEncoder constructs a DelegatingEncoder that encodes new
// passwords with the given primary Method. Verify and UpgradeNeeded dispatch
// via ByPrefix against the package method registry.
func NewDelegatingEncoder(primary *Method) *DelegatingEncoder {
	cassert.NotNil(primary, "primary method is nil")
	return &DelegatingEncoder{primary: primary}
}

// Encode delegates to the configured primary Method.
func (d *DelegatingEncoder) Encode(rawPassword string) (string, error) {
	cassert.NotNil(d, "delegating encoder is nil")
	cassert.NotNil(d.primary, "delegating encoder primary is nil")

	encoded, err := d.primary.Encode(rawPassword)
	if err != nil {
		return "", ErrDelegate(err)
	}
	return encoded, nil
}

// Verify resolves the encoding algorithm from the encoded password's prefix
// using ByPrefix and delegates verification to the matched Method. If no
// method matches the prefix, ErrUnknownEncodingPrefix is returned wrapped in
// a domain Error.
func (d *DelegatingEncoder) Verify(encodedPassword string, rawPassword string) (bool, error) {
	cassert.NotNil(d, "delegating encoder is nil")
	cassert.NotNil(d.primary, "delegating encoder primary is nil")

	method, err := ByPrefix(encodedPassword)
	if err != nil {
		return false, ErrDelegate(ErrUnknownEncodingPrefix, err)
	}

	ok, err := method.Verify(encodedPassword, rawPassword)
	if err != nil {
		return false, ErrDelegate(err)
	}
	return ok, nil
}

// UpgradeNeeded reports whether the encoded password should be re-encoded
// under the current primary. It returns (true, nil) whenever the encoded
// prefix resolves to a method whose Name() differs from primary.Name() —
// that is, the hash uses a legacy algorithm and should migrate. When the
// resolved method matches the primary, the call is delegated to the
// underlying Method.UpgradeNeeded for in-algorithm parameter drift checks.
func (d *DelegatingEncoder) UpgradeNeeded(encodedPassword string) (bool, error) {
	cassert.NotNil(d, "delegating encoder is nil")
	cassert.NotNil(d.primary, "delegating encoder primary is nil")

	method, err := ByPrefix(encodedPassword)
	if err != nil {
		return false, ErrDelegate(ErrUnknownEncodingPrefix, err)
	}

	if method.Name() != d.primary.Name() {
		return true, nil
	}

	needed, err := d.primary.UpgradeNeeded(encodedPassword)
	if err != nil {
		return false, ErrDelegate(err)
	}
	return needed, nil
}
