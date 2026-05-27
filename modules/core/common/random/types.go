// Package random provides fast non-secure random generation backed by math/rand/v2.
// For cryptographically secure randomness, use common/crypto/random instead.
package random

import ctypes "github.com/guidomantilla/yarumo/core/common/types"

var (
	_ BytesFn  = Bytes
	_ NumberFn = Number
	_ StringFn = String
	_ TextFn   = TextLower
	_ TextFn   = TextUpper
	_ TextFn   = TextNumber
	_ TextFn   = TextSpecial
	_ TextFn   = TextAlpha
	_ TextFn   = TextAlphaNum
	_ TextFn   = TextAll
)

// BytesFn is the function type for Bytes.
type BytesFn func(size int) ctypes.Bytes

// NumberFn is the function type for Number.
type NumberFn func(limit int64) int64

// StringFn is the function type for String.
type StringFn func(size int, charset string) string

// TextFn is the function type for convenience text generation functions.
type TextFn func(size int) string
