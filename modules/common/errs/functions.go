package errs

import (
	"errors"
	"reflect"
	"strings"

	cpointer "github.com/guidomantilla/yarumo/common/pointer"
	cutils "github.com/guidomantilla/yarumo/common/utils"
)

// As returns the error as a specific type T if it can be cast, otherwise returns a zero value of T and false.
func As[T error](err error) (T, bool) {
	target := cpointer.Zero[T]()
	ok := errors.As(err, &target)

	return target, ok
}

// Match checks if the error matches a specific type T or any of the provided values.
func Match[T error](err error, values ...error) bool {
	var target T
	if errors.As(err, &target) {
		return true
	}

	for _, v := range values {
		if errors.Is(err, v) {
			return true
		}
	}

	return false
}

// Wrap returns a new error that wraps the provided errors.
func Wrap(errs ...error) error {
	return errors.Join(errs...)
}

// Unwrap returns a slice of leaf errors by recursively unwrapping the provided error.
func Unwrap(err error) []error {
	seen := map[error]struct{}{}

	var out []error

	var walk func(error)

	walk = func(err error) {
		if cutils.Nil(err) {
			return
		}

		_, ok := seen[err]
		if ok {
			return
		}

		seen[err] = struct{}{}

		switch e := err.(type) {
		case interface{ Unwrap() []error }:
			for _, inner := range e.Unwrap() {
				walk(inner)
			}
		case interface{ Unwrap() error }:
			walk(e.Unwrap())
		default:
			out = append(out, err)
		}
	}
	walk(err)

	return out
}

// ErrorMessages returns a slice of error messages from the provided error by unwrapping it.
func ErrorMessages(err error) []string {
	unwrapped := Unwrap(err)
	if len(unwrapped) == 0 {
		return nil
	}

	msgs := make([]string, 0, len(unwrapped))
	for _, e := range unwrapped {
		msgs = append(msgs, e.Error())
	}

	return msgs
}

// HasErrorMessage checks if any of the unwrapped errors contain the specified substring in their error message.
func HasErrorMessage(err error, substr string) bool {
	for _, e := range Unwrap(err) {
		if strings.Contains(e.Error(), substr) {
			return true
		}
	}

	return false
}

// leafTypeName resolves the type name for a leaf error. If a currentType was propagated
// from a TypedError ancestor, it is used. Otherwise, falls back to the reflect type name.
func leafTypeName(err error, currentType string) string {
	if currentType != "" {
		return currentType
	}

	t := reflect.TypeOf(err)
	if t != nil {
		return t.String()
	}

	return "<nil>"
}

// AsErrorInfo converts an error tree into a slice of ErrorInfo, grouping leaf error messages
// by the TypedError.Type of their nearest ancestor. Errors without a TypedError ancestor
// fall back to the reflect type name.
func AsErrorInfo(err error) []ErrorInfo {
	if cutils.Nil(err) {
		return nil
	}

	grouped := make(map[string][]string)

	var order []string

	seen := map[error]struct{}{}

	var walk func(error, string)

	walk = func(err error, currentType string) {
		if cutils.Nil(err) {
			return
		}

		_, ok := seen[err]
		if ok {
			return
		}

		seen[err] = struct{}{}

		te, teOk := err.(interface{ ErrorType() string })
		if teOk {
			currentType = te.ErrorType()
		}

		switch e := err.(type) {
		case interface{ Unwrap() []error }:
			for _, inner := range e.Unwrap() {
				walk(inner, currentType)
			}
		case interface{ Unwrap() error }:
			walk(e.Unwrap(), currentType)
		default:
			typ := leafTypeName(err, currentType)

			_, exists := grouped[typ]
			if !exists {
				order = append(order, typ)
			}

			grouped[typ] = append(grouped[typ], err.Error())
		}
	}

	walk(err, "")

	infos := make([]ErrorInfo, 0, len(order))
	for _, typ := range order {
		infos = append(infos, ErrorInfo{
			Type:     typ,
			Messages: grouped[typ],
		})
	}

	return infos
}
