package passwords

import (
	rand "math/rand/v2"
	"strings"

	crandom "github.com/guidomantilla/yarumo/common/random"
)

// Generator generates and validates passwords.
type Generator struct {
	passwordLength int
	minSpecialChar int
	minNum         int
	minUpperCase   int
}

// NewGenerator creates a new password generator with the given options.
func NewGenerator(opts ...GeneratorOption) *Generator {
	options := NewGeneratorOptions(opts...)
	return &Generator{
		passwordLength: options.passwordLength,
		minSpecialChar: options.minSpecialChar,
		minNum:         options.minNum,
		minUpperCase:   options.minUpperCase,
	}
}

// Generate creates a random password meeting the configured requirements.
func (g *Generator) Generate() string {

	var password strings.Builder

	text, _ := crandom.TextSpecial(g.minSpecialChar)
	password.WriteString(text)

	text, _ = crandom.TextNumber(g.minNum)
	password.WriteString(text)

	text, _ = crandom.TextUpper(g.minUpperCase)
	password.WriteString(text)

	remainingLength := g.passwordLength - g.minSpecialChar - g.minNum - g.minUpperCase
	text, _ = crandom.TextLower(remainingLength)
	password.WriteString(text)

	inRune := []rune(password.String())
	rand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})
	return string(inRune)
}

const (
	specialCharSet = "@#$%^&*-_!+=[]{}|\\:',.?/`~\"();<>"
	numberSet      = "0123456789"
	upperCharSet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// Validate checks if a password meets the configured requirements.
func (g *Generator) Validate(rawPassword string) error {

	if len(rawPassword) < g.passwordLength {
		return ErrValidation(ErrPasswordLength)
	}

	specialCount := 0
	numCount := 0
	upperCount := 0

	for _, c := range rawPassword {
		switch {
		case strings.ContainsRune(specialCharSet, c):
			specialCount++
		case strings.ContainsRune(numberSet, c):
			numCount++
		case strings.ContainsRune(upperCharSet, c):
			upperCount++
		}
	}

	if specialCount < g.minSpecialChar {
		return ErrValidation(ErrPasswordSpecialChars)
	}

	if numCount < g.minNum {
		return ErrValidation(ErrPasswordNumbers)
	}

	if upperCount < g.minUpperCase {
		return ErrValidation(ErrPasswordUppercaseChars)
	}

	return nil
}
