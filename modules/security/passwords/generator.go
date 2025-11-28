package passwords

import (
	"math/rand/v2"
	"strings"

	"github.com/guidomantilla/yarumo/common/random"
)

const (
	lowerCharSet   = "abcdedfghijklmnopqrst"
	upperCharSet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialCharSet = "@#$%^&*-_!+=[]{}|\\:',.?/`~\"();<>"
	numberSet      = "0123456789"
	allCharSet     = lowerCharSet + upperCharSet + specialCharSet + numberSet
	//allCharSet = lowerCharSet + upperCharSet + numberSet
)

type generator struct {
	passwordLength int
	minSpecialChar int
	minNum         int
	minUpperCase   int
}

func NewGenerator(opts ...GeneratorOption) Generator {
	options := NewGeneratorOptions(opts...)
	return &generator{
		passwordLength: options.passwordLength,
		minSpecialChar: options.minSpecialChar,
		minNum:         options.minNum,
		minUpperCase:   options.minUpperCase,
	}
}

func (generator *generator) Generate() string {
	var password strings.Builder

	//Set special characters
	text, _ := random.TextSpecial(generator.minSpecialChar)
	password.WriteString(text)

	//Set numeric characters
	text, _ = random.TextNumber(generator.minNum)
	password.WriteString(text)

	//Set uppercase characters
	text, _ = random.TextUpper(generator.minUpperCase)
	password.WriteString(text)

	//Set lowercase characters
	remainingLength := generator.passwordLength - generator.minSpecialChar - generator.minNum - generator.minUpperCase
	text, _ = random.TextLower(remainingLength)
	password.WriteString(text)

	inRune := []rune(password.String())
	rand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})
	return string(inRune)
}

func (generator *generator) Validate(rawPassword string) error {

	if len(rawPassword) < generator.passwordLength {
		return ErrPasswordValidationFailed(ErrPasswordLength)
	}

	minSpecialCharCont := 0
	minNumCont := 0
	minUpperCaseCont := 0

	for _, c := range rawPassword {
		switch {
		case strings.ContainsRune(specialCharSet, c):
			minSpecialCharCont++
		case strings.ContainsRune(numberSet, c):
			minNumCont++
		case strings.ContainsRune(upperCharSet, c):
			minUpperCaseCont++
		}
	}

	if minSpecialCharCont < generator.minSpecialChar {
		return ErrPasswordValidationFailed(ErrPasswordSpecialChars)
	}

	if minNumCont < generator.minNum {
		return ErrPasswordValidationFailed(ErrPasswordNumbers)
	}

	if minUpperCaseCont < generator.minUpperCase {
		return ErrPasswordValidationFailed(ErrPasswordUppercaseChars)
	}

	return nil
}
