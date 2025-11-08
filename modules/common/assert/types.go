package assert

var (
	_ NotEmptyFn = NotEmpty
	_ NotNilFn   = NotNil
	_ EqualFn    = Equal
	_ NotEqualFn = NotEqual
	_ TrueFn     = True
	_ FalseFn    = False
)

type NotEmptyFn func(object any, message string)

type NotNilFn func(object any, message string)

type EqualFn func(val1 any, val2 any, message string)

type NotEqualFn func(val1 any, val2 any, message string)

type TrueFn func(condition bool, message string)

type FalseFn func(condition bool, message string)
