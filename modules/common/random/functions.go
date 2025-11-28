package random

import (
    crand "crypto/rand"
    "math/big"
    "strings"
)

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
var randInt = crand.Int

// Key returns cryptographically random bytes.
func Key(size int) []byte {
    key := make([]byte, size)
    _, _ = crand.Read(key)
    return key
}

// Number returns a cryptographically random integer in [0, max).
func Number(max int64) (int64, error) {
    n, err := randInt(crand.Reader, big.NewInt(max))
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

	for i := 0; i < size; i++ {
		random, err := Number(charsetLen)
		if err != nil {
			return "", err
		}
		out.WriteRune(charsetRunes[random])
	}

	return out.String(), nil
}

//
// Convenience functions
//

func TextLower(size int) (string, error) {
	return String(size, LowerCharSet)
}

func TextUpper(size int) (string, error) {
	return String(size, UpperCharSet)
}

func TextNumber(size int) (string, error) {
	return String(size, NumberSet)
}

func TextSpecial(size int) (string, error) {
	return String(size, SpecialCharSet)
}

func TextAlpha(size int) (string, error) {
	return String(size, AlphaSet)
}

func TextAlphaNum(size int) (string, error) {
	return String(size, AlphaNumSet)
}

func TextAll(size int) (string, error) {
	return String(size, AllCharSet)
}
