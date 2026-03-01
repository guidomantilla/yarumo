package random

import (
	"crypto/rand"
	"math/big"
	"strings"

	ctypes "github.com/guidomantilla/yarumo/common/types"
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

// randInt is an indirection to crypto/rand.Int to allow error-path testing.
// Tests may override this variable within the package to simulate failures.
var randInt = rand.Int

// Bytes returns cryptographically random bytes.
func Bytes(size int) ctypes.Bytes {
	if size <= 0 {
		return nil
	}

	key := make([]byte, size)
	_, _ = rand.Read(key)

	return key
}

// Number returns a cryptographically random integer in [0, max).
func Number(limit int64) (int64, error) {
	if limit <= 0 {
		return 0, nil
	}

	n, err := randInt(rand.Reader, big.NewInt(limit))
	if err != nil {
		return 0, err
	}

	return n.Int64(), nil
}

// String returns a cryptographically random string with custom charset.
func String(size int, charset string) (string, error) {
	if size <= 0 || len(charset) == 0 {
		return "", nil
	}

	charsetRunes := []rune(charset)
	charsetLen := int64(len(charsetRunes))

	var out strings.Builder
	out.Grow(size)

	for range size {
		random, err := Number(charsetLen)
		if err != nil {
			return "", err
		}

		out.WriteRune(charsetRunes[random])
	}

	return out.String(), nil
}

// --- Convenience functions ---

// TextLower returns a cryptographically random lowercase alphabetic string of the specified size.
func TextLower(size int) (string, error) {
	return String(size, LowerCharSet)
}

// TextUpper returns a cryptographically random uppercase alphabetic string of the specified size.
func TextUpper(size int) (string, error) {
	return String(size, UpperCharSet)
}

// TextNumber returns a cryptographically random numeric string of the specified size.
func TextNumber(size int) (string, error) {
	return String(size, NumberSet)
}

// TextSpecial returns a cryptographically random special character string of the specified size.
func TextSpecial(size int) (string, error) {
	return String(size, SpecialCharSet)
}

// TextAlpha returns a cryptographically random alphabetic string of the specified size.
func TextAlpha(size int) (string, error) {
	return String(size, AlphaSet)
}

// TextAlphaNum returns a cryptographically random alphanumeric string of the specified size.
func TextAlphaNum(size int) (string, error) {
	return String(size, AlphaNumSet)
}

// TextAll returns a cryptographically random string of the specified size.
func TextAll(size int) (string, error) {
	return String(size, AllCharSet)
}
