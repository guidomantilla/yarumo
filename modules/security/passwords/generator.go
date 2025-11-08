package passwords

import (
	"crypto/rand"
	"math/big"
	math "math/rand"
	"strings"
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

	//Set special character
	for i := 0; i < generator.minSpecialChar; i++ {
		random, _ := rand.Int(rand.Reader, big.NewInt(int64(len(specialCharSet))))
		password.WriteString(string(specialCharSet[random.Int64()]))
	}

	//Set numeric
	for i := 0; i < generator.minNum; i++ {
		random, _ := rand.Int(rand.Reader, big.NewInt(int64(len(numberSet))))
		password.WriteString(string(numberSet[random.Int64()]))
	}

	//Set uppercase
	for i := 0; i < generator.minUpperCase; i++ {
		random, _ := rand.Int(rand.Reader, big.NewInt(int64(len(upperCharSet))))
		password.WriteString(string(upperCharSet[random.Int64()]))
	}

	remainingLength := generator.passwordLength - generator.minSpecialChar - generator.minNum - generator.minUpperCase
	for i := 0; i < remainingLength; i++ {
		random, _ := rand.Int(rand.Reader, big.NewInt(int64(len(allCharSet))))
		password.WriteString(string(allCharSet[random.Int64()]))
	}
	inRune := []rune(password.String())
	math.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})
	return string(inRune)
}

func (generator *generator) Validate(rawPassword string) error {

	if len(rawPassword) < generator.passwordLength {
		return ErrPasswordLength
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
		return ErrPasswordSpecialChars
	}

	if minNumCont < generator.minNum {
		return ErrPasswordNumbers
	}

	if minUpperCaseCont < generator.minUpperCase {
		return ErrPasswordUppercaseChars
	}

	return nil
}
