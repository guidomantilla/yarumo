package random

import (
	rand "math/rand/v2"
	"strings"

	ctypes "github.com/guidomantilla/yarumo/core/common/types"
)

// Character sets for random string generation.
const (
	LowerCharSet   = "abcdefghijklmnopqrstuvwxyz"
	UpperCharSet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	NumberSet      = "0123456789"
	SpecialCharSet = "@#$%^&*-_!+=[]{}|\\:',.?/`~\"();<>"
	AlphaSet       = LowerCharSet + UpperCharSet
	AlphaNumSet    = AlphaSet + NumberSet
	AllCharSet     = AlphaNumSet + SpecialCharSet
)

// Bytes returns size random bytes using math/rand/v2.
// NOT cryptographically secure — use common/crypto/random for secrets, tokens,
// or any value that must be unpredictable.
func Bytes(size int) ctypes.Bytes {
	if size <= 0 {
		return nil
	}

	out := make([]byte, size)
	for i := range out {
		out[i] = byte(rand.Uint32())
	}

	return out
}

// Number returns a non-secure random integer in [0, limit).
// NOT cryptographically secure — use common/crypto/random for secrets, tokens,
// or any value that must be unpredictable.
func Number(limit int64) int64 {
	if limit <= 0 {
		return 0
	}

	return rand.Int64N(limit)
}

// String returns a non-secure random string with custom charset.
// NOT cryptographically secure — use common/crypto/random for secrets, tokens,
// or any value that must be unpredictable.
func String(size int, charset string) string {
	if size <= 0 || len(charset) == 0 {
		return ""
	}

	runes := []rune(charset)
	n := len(runes)

	var out strings.Builder
	out.Grow(size)

	for range size {
		out.WriteRune(runes[rand.IntN(n)])
	}

	return out.String()
}

// --- Convenience functions ---

// TextLower returns a non-secure random lowercase alphabetic string of the specified size.
// NOT cryptographically secure — use common/crypto/random for secrets, tokens,
// or any value that must be unpredictable.
func TextLower(size int) string {
	return String(size, LowerCharSet)
}

// TextUpper returns a non-secure random uppercase alphabetic string of the specified size.
// NOT cryptographically secure — use common/crypto/random for secrets, tokens,
// or any value that must be unpredictable.
func TextUpper(size int) string {
	return String(size, UpperCharSet)
}

// TextNumber returns a non-secure random numeric string of the specified size.
// NOT cryptographically secure — use common/crypto/random for secrets, tokens,
// or any value that must be unpredictable.
func TextNumber(size int) string {
	return String(size, NumberSet)
}

// TextSpecial returns a non-secure random special character string of the specified size.
// NOT cryptographically secure — use common/crypto/random for secrets, tokens,
// or any value that must be unpredictable.
func TextSpecial(size int) string {
	return String(size, SpecialCharSet)
}

// TextAlpha returns a non-secure random alphabetic string of the specified size.
// NOT cryptographically secure — use common/crypto/random for secrets, tokens,
// or any value that must be unpredictable.
func TextAlpha(size int) string {
	return String(size, AlphaSet)
}

// TextAlphaNum returns a non-secure random alphanumeric string of the specified size.
// NOT cryptographically secure — use common/crypto/random for secrets, tokens,
// or any value that must be unpredictable.
func TextAlphaNum(size int) string {
	return String(size, AlphaNumSet)
}

// TextAll returns a non-secure random string of the specified size.
// NOT cryptographically secure — use common/crypto/random for secrets, tokens,
// or any value that must be unpredictable.
func TextAll(size int) string {
	return String(size, AllCharSet)
}
