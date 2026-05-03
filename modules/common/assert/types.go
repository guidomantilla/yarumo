// Package assert provides runtime assertion functions that log errors or fatals
// depending on whether assertions are enabled.
package assert

var (
	_ NotEmptyFn = NotEmpty
	_ NotNilFn   = NotNil
	_ EqualFn    = Equal
	_ NotEqualFn = NotEqual
	_ TrueFn     = True
	_ FalseFn    = False
)

// NotEmptyFn is the function type for NotEmpty.
type NotEmptyFn func(object any, message string)

// NotNilFn is the function type for NotNil.
type NotNilFn func(object any, message string)

// EqualFn is the function type for Equal.
type EqualFn func(val1 any, val2 any, message string)

// NotEqualFn is the function type for NotEqual.
type NotEqualFn func(val1 any, val2 any, message string)

// TrueFn is the function type for True.
type TrueFn func(condition bool, message string)

// FalseFn is the function type for False.
type FalseFn func(condition bool, message string)
