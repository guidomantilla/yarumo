package utils

var (
	_ EqualFn    = Equal
	_ NotEqualFn = NotEqual
	_ NilFn      = Nil
	_ NotNilFn   = NotNil
	_ EmptyFn    = Empty
	_ NotEmptyFn = NotEmpty
)

type EqualFn func(x any, y any) bool

type NotEqualFn func(x any, y any) bool

type NilFn func(x any) bool

type NotNilFn func(x any) bool

type EmptyFn func(x any) bool

type NotEmptyFn func(x any) bool
