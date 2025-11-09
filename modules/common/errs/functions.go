package errs

import (
	"errors"
	"reflect"
	"strings"

	"github.com/guidomantilla/yarumo/common/pointer"
	"github.com/guidomantilla/yarumo/common/utils"
)

// As returns the error as a specific type T if it can be cast, otherwise returns a zero value of T and false.
func As[T error](err error) (T, bool) {
	target := pointer.Zero[T]()
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

// Unwrap returns a slice of errors by recursively unwrapping the provided error.
func Unwrap(err error) []error {
	seen := map[error]struct{}{}
	var out []error

	var walk func(error)
	walk = func(err error) {
		if utils.Nil(err) {
			return
		}
		if _, ok := seen[err]; ok {
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
	var msgs []string
	for _, e := range Unwrap(err) {
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

//

// AsErrorInfo converts an error into a slice of ErrorInfo, which contains the type and message of each unwrapped error.
func AsErrorInfo(err error) []ErrorInfo {
	var result []ErrorInfo

	for _, e := range Unwrap(err) {
		typ := reflect.TypeOf(e)
		typeName := "<nil>"
		if typ != nil {
			typeName = typ.String()
		}

		result = append(result, ErrorInfo{
			Type:    typeName,
			Message: e.Error(),
		})
	}

	return result
}
