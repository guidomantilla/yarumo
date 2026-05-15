package utils

// Character sets for random string generation.
const (
	LowerCaseLettersCharset = "abcdefghijklmnopqrstuvwxyz"
	UpperCaseLettersCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	LettersCharset          = LowerCaseLettersCharset + UpperCaseLettersCharset
	NumbersCharset          = "0123456789"
	AlphanumericCharset     = LettersCharset + NumbersCharset
	SpecialCharset          = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	AllCharset              = AlphanumericCharset + SpecialCharset
)
