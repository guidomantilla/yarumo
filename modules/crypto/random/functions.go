package random

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strings"

	ctypes "github.com/guidomantilla/yarumo/common/types"
)

// ErrShortRead is returned when crypto/rand.Read returns fewer bytes than
// requested. crypto/rand.Read should never return a short read on the
// supported platforms, but the contract allows it and Bytes refuses to
// silently return a partial buffer.
var ErrShortRead = errors.New("crypto/rand short read")

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

// randRead is an indirection to crypto/rand.Read to allow error-path testing.
// Tests may override this variable within the package to simulate failures.
var randRead = rand.Read

// Bytes returns cryptographically random bytes. It returns a nil slice with a
// nil error when size is non-positive, and a non-nil error when the underlying
// crypto/rand source fails to deliver the requested number of bytes.
func Bytes(size int) (ctypes.Bytes, error) {
	if size <= 0 {
		return nil, nil
	}

	key := make([]byte, size)

	n, err := randRead(key)
	if err != nil {
		return nil, err
	}

	if n != size {
		return nil, ErrShortRead
	}

	return key, nil
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
