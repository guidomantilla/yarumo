package errs

import (
	"fmt"
)

// errA is a simple custom error type
type errA struct{ msg string }

func (e *errA) Error() string { return e.msg }

// errB wraps a single inner error and exposes Unwrap() error
type errB struct{ inner error }

func (e *errB) Error() string { return "errB: " + fmt.Sprint(e.inner) }

func (e *errB) Unwrap() error { return e.inner }

// errMulti wraps multiple inner errors and exposes Unwrap() []error
type errMulti struct{ inners []error }

func (e *errMulti) Error() string { return "errMulti" }

func (e *errMulti) Unwrap() []error { return e.inners }

// cyc is used to validate that Unwrap() doesn't loop forever on cycles
type cyc struct{ inner error }

func (e *cyc) Error() string { return "cyc" }

func (e *cyc) Unwrap() error { return e.inner }
