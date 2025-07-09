package errs

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/guidomantilla/yarumo/pkg/common/pointer"
	"github.com/guidomantilla/yarumo/pkg/common/utils"
)

// As
func As[T error](err error) (T, bool) {
	target := pointer.Zero[T]()
	ok := errors.As(err, &target)
	return target, ok
}

// Match
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

// ------ //

type TypedError struct {
	Type string
	Err  error
}

func (e *TypedError) Error() string {
	return fmt.Sprintf("%s error: %s", e.Type, e.Err)
}

func (e *TypedError) Unwrap() error {
	return e.Err
}

// ------ //

// Unwrap
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

// ErrorMessages
func ErrorMessages(err error) []string {
	var msgs []string
	for _, e := range Unwrap(err) {
		msgs = append(msgs, e.Error())
	}
	return msgs
}

// HasErrorMessage
func HasErrorMessage(err error, substr string) bool {
	for _, e := range Unwrap(err) {
		if strings.Contains(e.Error(), substr) {
			return true
		}
	}
	return false
}

//

type ErrorInfo struct {
	Type    string `json:"type,omitempty"`
	Message string `json:"message"`
}

// AsErrorInfo
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
