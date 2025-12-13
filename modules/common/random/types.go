package random

import "github.com/guidomantilla/yarumo/common/types"

var (
	_ BytesFn  = Bytes
	_ NumberFn = Number
	_ StringFn = String
	_ TextFn   = TextLower
	_ TextFn   = TextUpper
	_ TextFn   = TextNumber
	_ TextFn   = TextSpecial
	_ TextFn   = TextAlpha
	_ TextFn   = TextAlphaNum
	_ TextFn   = TextAll
)

type BytesFn func(size int) types.Bytes

type NumberFn func(max int64) (int64, error)

type StringFn func(size int, charset string) (string, error)

type TextFn func(size int) (string, error)
