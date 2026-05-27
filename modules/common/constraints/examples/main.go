// Package main demonstrates the generic type constraints from
// common/constraints. The package only declares constraints (Signed,
// Unsigned, Integer, Float, Number, Complex, Ordenable, Comparable), so
// this demo defines a handful of generic helpers locally and instantiates
// them with multiple concrete types.
package main

import (
	"fmt"

	cconstraints "github.com/guidomantilla/yarumo/common/constraints"
)

// Max returns the larger of two ordered values.
func Max[T cconstraints.Ordenable](a, b T) T {
	if a > b {
		return a
	}

	return b
}

// SumNumbers sums any slice of integer or floating-point values.
func SumNumbers[T cconstraints.Number](values []T) T {
	var total T
	for _, v := range values {
		total += v
	}

	return total
}

// Negate flips the sign of any signed integer.
func Negate[T cconstraints.Signed](v T) T {
	return -v
}

// Distinct reports whether every element in slice is unique.
func Distinct[T cconstraints.Comparable](slice []T) bool {
	seen := map[T]struct{}{}
	for _, v := range slice {
		_, ok := seen[v]
		if ok {
			return false
		}

		seen[v] = struct{}{}
	}

	return true
}

func main() {
	fmt.Println("=== Ordenable: Max ===")
	fmt.Printf("  Max(3, 7)         -> %d\n", Max(3, 7))
	fmt.Printf("  Max(1.5, 0.5)     -> %g\n", Max(1.5, 0.5))
	fmt.Printf("  Max(\"abc\", \"xyz\") -> %q\n", Max("abc", "xyz"))

	fmt.Println("=== Number: SumNumbers ===")
	fmt.Printf("  SumNumbers([]int)     -> %d\n", SumNumbers([]int{1, 2, 3, 4}))
	fmt.Printf("  SumNumbers([]float64) -> %g\n", SumNumbers([]float64{0.1, 0.2, 0.3}))

	fmt.Println("=== Signed: Negate ===")
	fmt.Printf("  Negate(int8(10))  -> %d\n", Negate(int8(10)))
	fmt.Printf("  Negate(int64(-5)) -> %d\n", Negate(int64(-5)))

	fmt.Println("=== Comparable: Distinct ===")
	fmt.Printf("  Distinct([]int{1, 2, 3}) -> %v\n", Distinct([]int{1, 2, 3}))
	fmt.Printf("  Distinct([]int{1, 2, 1}) -> %v\n", Distinct([]int{1, 2, 1}))
}
