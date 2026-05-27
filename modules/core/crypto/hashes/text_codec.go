package hashes

import (
	"encoding"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
)

// Type compliance.
var (
	_ encoding.TextMarshaler   = (*Method)(nil)
	_ encoding.TextUnmarshaler = (*Method)(nil)
)

// MarshalText implements encoding.TextMarshaler. It returns the registry name
// of this Method so callers can serialize algorithm choices into config files
// (YAML/JSON/TOML) and other text-oriented encodings.
//
// The receiver must be non-nil; callers that hold a nil *Method will trigger
// the cassert.NotNil invariant (and a nil pointer dereference) — Method
// values are intended to be obtained either from the predefined package
// variables (e.g. SHA256) or from Get.
func (m *Method) MarshalText() ([]byte, error) {
	cassert.NotNil(m, "method is nil")

	return []byte(m.name), nil
}

// UnmarshalText implements encoding.TextUnmarshaler. It looks the provided
// name up in the package registry via Get and overwrites the receiver with
// the resolved Method.
//
// The receiver must be a pre-allocated, non-nil *Method — the typical use is
// a struct field declared as Algorithm *hashes.Method which the decoder
// (encoding/json, viper, koanf, ...) allocates before calling UnmarshalText.
//
// Caveat: resolution happens against whatever the registry contains at the
// moment of the call. Custom methods registered after config load will not
// resolve here; callers that need late-bound lookup should call Get(name)
// directly.
func (m *Method) UnmarshalText(data []byte) error {
	cassert.NotNil(m, "method is nil")

	found, err := Get(string(data))
	if err != nil {
		return err
	}

	*m = *found

	return nil
}
