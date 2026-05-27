// Package main demonstrates the runtime assertion helpers from common/assert.
// Assertions are kept in log-only mode (Enable(false)) so the binary does not
// terminate on a failing check. The default Logger is a noop, so error lines
// are discarded — what you see printed here is the demo's own commentary.
package main

import (
	"fmt"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
)

func main() {
	// Keep assertions in log-only mode so a failing check does not kill the
	// process. With assertions disabled, NotNil/NotEmpty/Equal/True/False all
	// log at error level instead of calling clog.Fatal.
	cassert.Enable(false)

	demoNotNil()
	demoNotEmpty()
	demoEqual()
	demoTrueFalse()
}

// demoNotNil exercises NotNil with a non-nil and a nil receiver.
func demoNotNil() {
	fmt.Println("=== NotNil ===")

	value := "hello"
	cassert.NotNil(value, "value should not be nil")
	fmt.Println("  NotNil(\"hello\") passed")

	var nilPtr *int
	cassert.NotNil(nilPtr, "nilPtr should not be nil (this logs an error)")
	fmt.Println("  NotNil(nilPtr) logged an assertion failure")
}

// demoNotEmpty exercises NotEmpty with a populated and an empty slice.
func demoNotEmpty() {
	fmt.Println("=== NotEmpty ===")

	cassert.NotEmpty([]int{1, 2, 3}, "slice should not be empty")
	fmt.Println("  NotEmpty([1 2 3]) passed")

	cassert.NotEmpty("", "string should not be empty (this logs an error)")
	fmt.Println("  NotEmpty(\"\") logged an assertion failure")
}

// demoEqual exercises Equal and NotEqual.
func demoEqual() {
	fmt.Println("=== Equal / NotEqual ===")

	cassert.Equal(2+2, 4, "math should still work")
	fmt.Println("  Equal(2+2, 4) passed")

	cassert.NotEqual("yes", "no", "yes should not equal no")
	fmt.Println("  NotEqual(\"yes\", \"no\") passed")
}

// demoTrueFalse exercises True and False.
func demoTrueFalse() {
	fmt.Println("=== True / False ===")

	cassert.True(1 < 2, "1 should be less than 2")
	fmt.Println("  True(1 < 2) passed")

	cassert.False(1 > 2, "1 should not be greater than 2")
	fmt.Println("  False(1 > 2) passed")
}
