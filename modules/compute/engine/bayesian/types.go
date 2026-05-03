package bayesian

import "github.com/guidomantilla/yarumo/compute/math/stats"

// CPT is a conditional probability table.
// Maps parent configurations to child outcome distributions.
type CPT struct {
	Variable stats.Var
	Parents  []stats.Var
	Entries  map[string]stats.Distribution // serialized parent config -> Distribution.
}

// Factor is an intermediate table used in variable elimination.
type Factor struct {
	Variables []stats.Var
	Table     map[string]stats.Prob // serialized assignment -> probability.
}
