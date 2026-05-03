// Package adapters converts repository config types to inference engine types.
package adapters

import (
	"github.com/guidomantilla/yarumo/compute/math/logic"
)

// ParsedRule holds a parsed deductive rule definition.
type ParsedRule struct {
	Name       string
	Formula    logic.Formula
	Conclusion map[logic.Var]bool
	Priority   int
}

// MembershipParamCounts maps membership function type names to their expected parameter counts.
var MembershipParamCounts = map[string]int{
	"triangular":  3,
	"trapezoidal": 4,
	"gaussian":    2,
}
