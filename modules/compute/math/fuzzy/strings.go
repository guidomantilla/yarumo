package fuzzy

import "strconv"

// String returns a string representation of a fuzzy degree.
func (d Degree) String() string {
	return strconv.FormatFloat(float64(d), 'f', -1, 64)
}

// String returns a string representation of a fuzzy set.
func (s Set) String() string {
	return "Set(" + s.Name + ")"
}
