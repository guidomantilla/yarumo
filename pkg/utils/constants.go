package utils

const (
	LowerCaseLettersCharset = "abcdefghijklmnopqrstuvwxyz"
	UpperCaseLettersCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	LettersCharset          = LowerCaseLettersCharset + UpperCaseLettersCharset
	NumbersCharset          = "0123456789"
	AlphanumericCharset     = LettersCharset + NumbersCharset
	SpecialCharset          = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	AllCharset              = AlphanumericCharset + SpecialCharset
)
