package main

import (
	"fmt"
)

func main() {

	fmt.Println(min(1, 2))

	var sf superFloat = 0.3
	fmt.Println(minFloat64(sf, 6.3))

	fmt.Println(minTypes(sf, 2))
	fmt.Println(minTypes(sf, 5.6))
	fmt.Println(minTypes(1, 1))

	// Initialize a map for the integer values
	ints := map[string]int64{
		"first":  34,
		"second": 12,
	}

	// Initialize a map for the float values
	floats := map[string]float64{
		"first":  35.98,
		"second": 26.99,
	}

	fmt.Printf("Generic Sums: %v and %v\n",
		SumIntsOrFloats[string, int64](ints),
		SumIntsOrFloats[string, float64](floats))

	fmt.Printf("Generic Sums, type parameters inferred: %v and %v\n",
		SumIntsOrFloats(ints),
		SumIntsOrFloats(floats))

	fmt.Printf("Generic Sums with Constraint: %v and %v\n",
		SumNumbers(ints),
		SumNumbers(floats))
}

func min[T any](a T, b T) T {
	return a
}

type superFloat float64

func minFloat64[T ~float64](a T, b T) T {
	if a < b {
		return a
	}
	return b
}

type allowedTypes interface {
	~float64 | int
}

func minTypes[T allowedTypes](a T, b T) T {

	if a < b {
		return a
	}
	return b
}

type Number interface {
	int64 | float64
}

func SumIntsOrFloats[K comparable, V int64 | float64](m map[K]V) V {
	var s V
	for _, v := range m {
		s += v
	}
	return s
}

func SumNumbers[K comparable, V Number](m map[K]V) V {
	var s V
	for _, v := range m {
		s += v
	}
	return s
}
