// Package generator produces random passwords that satisfy configurable
// character-class constraints (minimum special / numeric / uppercase /
// lowercase characters and total length), and validates raw passwords
// against the same constraints.
//
// # Placement
//
// The package lives at common/crypto/passwords/generator/ — a sibling of
// the password hasher at common/crypto/passwords/. The two packages are
// independent: this one synthesises passwords, the parent one stores them.
// Generator imports common/random for its crypto/rand-backed alphabet
// primitives; there is no import cycle (random has no upward dependencies).
//
// # Security model
//
// The generator is CSPRNG-backed end-to-end. Character selection uses
// common/random.Text* helpers (crypto/rand via math/big.Int.SetBits), and
// the final shuffle is a Fisher-Yates pass driven by common/random.Number
// — never math/rand. Treating the generator output as unpredictable
// therefore depends only on the platform crypto/rand entropy source.
//
// Generated passwords are intended to be paired with a password hasher on
// storage (see common/crypto/passwords). The generator alone does not
// provide replay protection, rotation, or storage hardening — it only
// guarantees that each emitted password is drawn uniformly at random from
// the configured constraint space.
//
// # Anti-patterns avoided
//
// This implementation deliberately addresses issues documented in the
// archived legacy package (gist cf0b78c1acb1ca704cd4e40e33788473):
//
//   - No math/rand shuffle — Fisher-Yates uses common/random.Number.
//   - No silent option rejection — With<Field> options accept any value;
//     NewGenerator returns a typed error when the sum of minimums exceeds
//     the total length.
//   - No validation/option threshold duplication — the validator checks
//     against the configured Options, not against package-level constants.
//   - Explicit minimum-lowercase knob (WithMinLowerCase).
package generator

import (
	"strings"

	cassert "github.com/guidomantilla/yarumo/common/assert"
)

// Character sets used to compose generated passwords. They mirror the
// alphabets exposed by common/random and are duplicated here as private
// constants only to keep Validate independent of import-time changes.
const (
	lowerCharSet   = "abcdefghijklmnopqrstuvwxyz"
	upperCharSet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberSet      = "0123456789"
	specialCharSet = "@#$%^&*-_!+=[]{}|\\:',.?/`~\"();<>"
)

// Generator generates and validates passwords against a fixed set of
// character-class constraints captured at construction time.
type Generator struct {
	passwordLength int
	minSpecialChar int
	minNumber      int
	minUpperCase   int
	minLowerCase   int
}

// NewGenerator creates a new password generator from the given options.
//
// It returns an error when the configuration is internally inconsistent —
// specifically, when the sum of character-class minimums exceeds the
// configured total password length. This replaces the legacy package's
// silent option rejection with an explicit failure mode the caller can
// react to (e.g., surface a config error at startup).
func NewGenerator(opts ...Option) (*Generator, error) {
	options := NewOptions(opts...)

	sum := options.minSpecialChar + options.minNumber + options.minUpperCase + options.minLowerCase
	if sum > options.passwordLength {
		return nil, ErrConfiguration(ErrConstraintsExceedLength)
	}

	return &Generator{
		passwordLength: options.passwordLength,
		minSpecialChar: options.minSpecialChar,
		minNumber:      options.minNumber,
		minUpperCase:   options.minUpperCase,
		minLowerCase:   options.minLowerCase,
	}, nil
}

// Generate creates a random password meeting the configured constraints.
//
// The result is composed by drawing the required quota of special,
// numeric, uppercase and lowercase characters and then filling the
// remaining slots from the lowercase alphabet. The combined string is
// shuffled with a Fisher-Yates pass backed by crypto/rand so the position
// of each character class is unpredictable.
func (g *Generator) Generate() (string, error) {
	cassert.NotNil(g, "generator is nil")

	return generate(g)
}

// Validate checks whether rawPassword satisfies the generator's configured
// constraints. It is the inverse of Generate: any password produced by the
// same Generator must pass Validate.
//
// Returns nil on success, or an *Error wrapping the first constraint that
// failed (length first, then special, number, uppercase, lowercase).
func (g *Generator) Validate(rawPassword string) error {
	cassert.NotNil(g, "generator is nil")

	return validate(g, rawPassword)
}

// PasswordLength returns the configured total password length.
func (g *Generator) PasswordLength() int {
	cassert.NotNil(g, "generator is nil")
	return g.passwordLength
}

// MinSpecialChar returns the configured minimum number of special characters.
func (g *Generator) MinSpecialChar() int {
	cassert.NotNil(g, "generator is nil")
	return g.minSpecialChar
}

// MinNumber returns the configured minimum number of numeric characters.
func (g *Generator) MinNumber() int {
	cassert.NotNil(g, "generator is nil")
	return g.minNumber
}

// MinUpperCase returns the configured minimum number of uppercase characters.
func (g *Generator) MinUpperCase() int {
	cassert.NotNil(g, "generator is nil")
	return g.minUpperCase
}

// MinLowerCase returns the configured minimum number of lowercase characters.
func (g *Generator) MinLowerCase() int {
	cassert.NotNil(g, "generator is nil")
	return g.minLowerCase
}

// classCounts tallies characters by class for Validate.
type classCounts struct {
	special int
	number  int
	upper   int
	lower   int
}

// countClasses counts characters per character class in a single pass.
func countClasses(rawPassword string) classCounts {
	var counts classCounts
	for _, c := range rawPassword {
		switch {
		case strings.ContainsRune(specialCharSet, c):
			counts.special++
		case strings.ContainsRune(numberSet, c):
			counts.number++
		case strings.ContainsRune(upperCharSet, c):
			counts.upper++
		case strings.ContainsRune(lowerCharSet, c):
			counts.lower++
		}
	}
	return counts
}
