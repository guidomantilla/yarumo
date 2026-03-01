// Package errs provides typed error handling utilities, error chain manipulation,
// and JSON-serializable error representation.
package errs

var (
	_ error             = (*TypedError)(nil)
	_ AsFn[error]       = As[error]
	_ MatchFn[error]    = Match[error]
	_ WrapFn            = Wrap
	_ UnwrapFn          = Unwrap
	_ ErrorMessagesFn   = ErrorMessages
	_ HasErrorMessageFn = HasErrorMessage
	_ AsErrorInfoFn     = AsErrorInfo
)

// AsFn is the function type for As.
type AsFn[T error] func(err error) (T, bool)

// MatchFn is the function type for Match.
type MatchFn[T error] func(err error, values ...error) bool

// WrapFn is the function type for Wrap.
type WrapFn func(errs ...error) error

// UnwrapFn is the function type for Unwrap.
type UnwrapFn func(err error) []error

// ErrorMessagesFn is the function type for ErrorMessages.
type ErrorMessagesFn func(err error) []string

// HasErrorMessageFn is the function type for HasErrorMessage.
type HasErrorMessageFn func(err error, substr string) bool

// AsErrorInfoFn is the function type for AsErrorInfo.
type AsErrorInfoFn func(err error) []ErrorInfo
