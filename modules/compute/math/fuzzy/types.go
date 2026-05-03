// Package fuzzy provides fuzzy logic primitives.
package fuzzy

// Degree is a fuzzy truth value in [0,1].
type Degree float64

// MembershipFn maps a crisp value to a fuzzy degree.
type MembershipFn func(x float64) Degree

// TNormFn combines two fuzzy degrees (intersection).
type TNormFn func(a, b Degree) Degree

// TConormFn combines two fuzzy degrees (union).
type TConormFn func(a, b Degree) Degree

// DefuzzifyFn converts sampled fuzzy output to a crisp value.
// xs = domain points, ys = membership degrees at each point.
type DefuzzifyFn func(xs []float64, ys []Degree) float64

// Set is a named fuzzy set with a membership function.
type Set struct {
	Name string
	Fn   MembershipFn
}

// Point is an (x, degree) pair used in discretized fuzzy sets.
type Point struct {
	X      float64
	Degree Degree
}
