package generator

import (
	"strings"

	crandom "github.com/guidomantilla/yarumo/core/crypto/random"
)

// Package-level indirections to the crypto/rand-backed alphabet helpers in
// common/crypto/random. Tests may override these to exercise error paths
// from the upstream CSPRNG without forking the dependency. The same pattern
// is used by common/crypto/random.randInt.
var (
	textSpecial = crandom.TextSpecial
	textNumber  = crandom.TextNumber
	textUpper   = crandom.TextUpper
	textLower   = crandom.TextLower
	randNumber  = crandom.Number
)

// generate composes a password meeting the generator's constraints.
//
// Defensive validation lives here, per the workspace convention that
// private helpers taking a *Generator start with a nil-receiver guard.
// Counter-intuitively, the *Generator value reaches this function from
// the exported Method receiver, which already cassert.NotNil-s; we still
// double-check for the indirect call paths (fuzz harness, future callers).
func generate(g *Generator) (string, error) {
	if g == nil {
		return "", ErrGeneration(ErrGeneratorIsNil)
	}

	var password strings.Builder
	password.Grow(g.passwordLength)

	text, err := textSpecial(g.minSpecialChar)
	if err != nil {
		return "", ErrGeneration(err)
	}
	password.WriteString(text)

	text, err = textNumber(g.minNumber)
	if err != nil {
		return "", ErrGeneration(err)
	}
	password.WriteString(text)

	text, err = textUpper(g.minUpperCase)
	if err != nil {
		return "", ErrGeneration(err)
	}
	password.WriteString(text)

	text, err = textLower(g.minLowerCase)
	if err != nil {
		return "", ErrGeneration(err)
	}
	password.WriteString(text)

	// Fill the remaining slots with lowercase letters. NewGenerator's
	// constraint check guarantees this value is >= 0.
	remaining := g.passwordLength - g.minSpecialChar - g.minNumber - g.minUpperCase - g.minLowerCase
	text, err = textLower(remaining)
	if err != nil {
		return "", ErrGeneration(err)
	}
	password.WriteString(text)

	runes := []rune(password.String())
	errShuffle := shuffleRunes(runes)
	if errShuffle != nil {
		// shuffleRunes already returns a domain *Error; propagate as-is.
		return "", errShuffle
	}

	return string(runes), nil
}

// validate enforces the generator's constraints against rawPassword. The
// caller-supplied raw password is treated as a runtime input; per the
// workspace convention we use a sentinel error rather than cassert.
func validate(g *Generator, rawPassword string) error {
	if g == nil {
		return ErrValidation(ErrGeneratorIsNil)
	}

	if len(rawPassword) < g.passwordLength {
		return ErrValidation(ErrPasswordLength)
	}

	counts := countClasses(rawPassword)

	if counts.special < g.minSpecialChar {
		return ErrValidation(ErrPasswordSpecialChars)
	}
	if counts.number < g.minNumber {
		return ErrValidation(ErrPasswordNumbers)
	}
	if counts.upper < g.minUpperCase {
		return ErrValidation(ErrPasswordUppercaseChars)
	}
	if counts.lower < g.minLowerCase {
		return ErrValidation(ErrPasswordLowercaseChars)
	}

	return nil
}

// shuffleRunes performs an in-place Fisher-Yates shuffle backed by
// crypto/rand via common/crypto/random.Number. It replaces the legacy code's
// math/rand/v2.Shuffle, which produced predictable orderings unsuitable
// for a security-sensitive generator.
//
// Empty / single-element slices are no-ops. On crypto/rand failure the
// function returns an *Error wrapping ErrShuffleFailed without leaving
// the slice in a partially mutated state.
func shuffleRunes(runes []rune) error {
	for i := len(runes) - 1; i > 0; i-- {
		j, err := randNumber(int64(i + 1))
		if err != nil {
			return ErrGeneration(ErrShuffleFailed, err)
		}
		runes[i], runes[j] = runes[j], runes[i]
	}
	return nil
}
