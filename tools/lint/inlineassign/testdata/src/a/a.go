// Package a contains deliberate violations and one compliant counterexample to
// drive analysistest.Run for the inlineassign analyzer.
package a

import "errors"

// doSomething simulates a function that returns an error.
func doSomething() error { return errors.New("boom") }

// errCheck demonstrates the forbidden error-check form.
func errCheck() {
	if err := doSomething(); err != nil { // want `inline assignment in if-statement; split the assignment into its own statement before the if \(No Inline Assignments rule\)`
		_ = err
	}
}

// mapLookup demonstrates the forbidden map-lookup form.
func mapLookup() {
	m := map[string]int{"a": 1}
	if v, ok := m["a"]; ok { // want `inline assignment in if-statement; split the assignment into its own statement before the if \(No Inline Assignments rule\)`
		_ = v
	}
}

// typeAssert demonstrates the forbidden type-assertion form.
func typeAssert(x any) {
	if v, ok := x.(string); ok { // want `inline assignment in if-statement; split the assignment into its own statement before the if \(No Inline Assignments rule\)`
		_ = v
	}
}

// arbitraryInit demonstrates that any non-nil Init is forbidden, not just the
// three canonical forms.
func arbitraryInit() {
	if x := 1 + 1; x == 2 { // want `inline assignment in if-statement; split the assignment into its own statement before the if \(No Inline Assignments rule\)`
		_ = x
	}
}

// compliant is the counterexample: assignment split from condition.
func compliant() {
	err := doSomething()
	if err != nil {
		_ = err
	}

	m := map[string]int{"a": 1}
	v, ok := m["a"]
	if ok {
		_ = v
	}
}
