package assert

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	clog "github.com/guidomantilla/yarumo/common/log"
	cutils "github.com/guidomantilla/yarumo/common/utils"
)

var enabled atomic.Bool

// Enable toggles whether assertions log fatal errors or regular errors.
func Enable(v bool) {
	enabled.Store(v)
}

// NotEmpty checks if the object is not empty and logs a fatal error if it is.
func NotEmpty(object any, message string) {
	if cutils.Empty(object) {
		err := errors.New("assertion failed: object is empty")
		if enabled.Load() {
			clog.Fatal(context.Background(), message, "error", err)

			return
		}

		clog.Error(context.Background(), message, "error", err)
	}
}

// NotNil checks if the object is not nil and logs a fatal error if it is.
func NotNil(object any, message string) {
	if cutils.Nil(object) {
		err := errors.New("assertion failed: object is nil")
		if enabled.Load() {
			clog.Fatal(context.Background(), message, "error", err)

			return
		}

		clog.Error(context.Background(), message, "error", err)
	}
}

// Equal checks if two values are equal and logs a fatal error if they are not.
func Equal(val1 any, val2 any, message string) {
	if cutils.NotEqual(val1, val2) {
		err := fmt.Errorf("assertion failed: %v != %v", val1, val2)
		if enabled.Load() {
			clog.Fatal(context.Background(), message, "error", err)

			return
		}

		clog.Error(context.Background(), message, "error", err)
	}
}

// NotEqual checks if two values are not equal and logs a fatal error if they are.
func NotEqual(val1 any, val2 any, message string) {
	if cutils.Equal(val1, val2) {
		err := fmt.Errorf("assertion failed: %v == %v", val1, val2)
		if enabled.Load() {
			clog.Fatal(context.Background(), message, "error", err)

			return
		}

		clog.Error(context.Background(), message, "error", err)
	}
}

// True checks if a condition is true and logs a fatal error if it is not.
func True(condition bool, message string) {
	if !condition {
		err := errors.New("assertion failed: condition is false")
		if enabled.Load() {
			clog.Fatal(context.Background(), message, "error", err)

			return
		}

		clog.Error(context.Background(), message, "error", err)
	}
}

// False checks if a condition is false and logs a fatal error if it is not.
func False(condition bool, message string) {
	if condition {
		err := errors.New("assertion failed: condition is true")
		if enabled.Load() {
			clog.Fatal(context.Background(), message, "error", err)

			return
		}

		clog.Error(context.Background(), message, "error", err)
	}
}
