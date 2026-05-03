// Package constraints provides generic type constraints for use in type-parameterized functions.
package constraints

import "cmp"

// Signed is a constraint that permits any signed integer type.
// If future releases of Go add new predeclared signed integer types,
// this constraint will be modified to include them.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Unsigned is a constraint that permits any unsigned integer type.
// If future releases of Go add new predeclared unsigned integer types,
// this constraint will be modified to include them.
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Integer is a constraint that permits any integer type.
// If future releases of Go add new predeclared integer types,
// this constraint will be modified to include them.
type Integer interface {
	Signed | Unsigned
}

// Float is a constraint that permits any floating-point type.
// If future releases of Go add new predeclared floating-point types,
// this constraint will be modified to include them.
type Float interface {
	~float32 | ~float64
}

// Complex is a constraint that permits any complex numeric type.
// If future releases of Go add new predeclared complex numeric types,
// this constraint will be modified to include them.
type Complex interface {
	~complex64 | ~complex128
}

// --- Aliases ---

// Comparable is an alias for the built-in comparable constraint.
type Comparable = comparable

// Ordenable is an alias for cmp.Ordered, permitting any type that supports ordering operators.
type Ordenable = cmp.Ordered

// Number is a constraint that permits any integer or floating-point type.
// If future releases of Go add new predeclared numeric types,
// this constraint will be modified to include them.
type Number interface {
	Integer | Float
}
