// Package main demonstrates the typed-error utilities in common/errs:
// TypedError as the embeddable base, Wrap/Unwrap to compose error trees,
// Match/As to query them, and AsErrorInfo to extract a serialization-
// friendly view of the tree grouped by type tag.
package main

import (
	"errors"
	"fmt"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
)

// ValidationError is a sample domain error that embeds TypedError.
type ValidationError struct {
	cerrs.TypedError
}

// NewValidationError joins causes under the "validation" type tag.
func NewValidationError(causes ...error) error {
	return &ValidationError{
		TypedError: cerrs.TypedError{
			Type: "validation",
			Err:  errors.Join(causes...),
		},
	}
}

var (
	errFieldRequired = errors.New("field is required")
	errFieldTooLong  = errors.New("field too long")
)

func main() {
	demoTypedError()
	demoWrapMatch()
	demoAs()
	demoAsErrorInfo()
}

// demoTypedError builds a TypedError and prints the tagged Error() output.
func demoTypedError() {
	fmt.Println("=== TypedError ===")

	err := cerrs.NewTypedError("io", errors.New("disk full"))
	fmt.Printf("  %s\n", err)
}

// demoWrapMatch joins two sentinels into a tree and queries it via Match.
func demoWrapMatch() {
	fmt.Println("=== Wrap / Match ===")

	tree := cerrs.Wrap(errFieldRequired, errFieldTooLong)
	fmt.Printf("  joined: %s\n", tree)

	fmt.Printf("  Match[ValidationError](tree) -> %v\n", cerrs.Match[*ValidationError](tree))
	fmt.Printf("  Match(tree, errFieldRequired) -> %v\n", cerrs.Match[*ValidationError](tree, errFieldRequired))
}

// demoAs walks the tree looking for a domain-typed error.
func demoAs() {
	fmt.Println("=== As ===")

	wrapped := NewValidationError(errFieldRequired)
	tree := cerrs.Wrap(errors.New("outer"), wrapped)

	got, ok := cerrs.As[*ValidationError](tree)
	fmt.Printf("  As[*ValidationError] -> ok=%v, type=%q\n", ok, got.Type)
}

// demoAsErrorInfo flattens an error tree into the serializable ErrorInfo slice.
func demoAsErrorInfo() {
	fmt.Println("=== AsErrorInfo ===")

	tree := NewValidationError(errFieldRequired, errFieldTooLong)
	for _, info := range cerrs.AsErrorInfo(tree) {
		fmt.Printf("  type=%s\n", info.Type)

		for _, msg := range info.Messages {
			fmt.Printf("    - %s\n", msg)
		}
	}
}
