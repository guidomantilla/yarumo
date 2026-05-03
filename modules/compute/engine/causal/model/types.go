// Package model provides the structural causal model (SCM) for causal inference.
package model

// EquationFn computes a variable's value from its parent values.
type EquationFn func(parents map[string]float64) float64

// Variable represents a node in the structural causal model.
type Variable struct {
	Name     string
	Parents  []string
	Equation EquationFn
}

// SCM defines the structural causal model interface.
type SCM interface {
	// AddVariable adds a variable with its parents and structural equation.
	AddVariable(name string, parents []string, equation EquationFn) error
	// Variable returns the variable definition.
	Variable(name string) (Variable, bool)
	// Variables returns all variable names in topological order.
	Variables() []string
	// Parents returns the parent variable names.
	Parents(name string) []string
	// Children returns the child variable names.
	Children(name string) []string
	// Validate checks the model for cycles and missing parents.
	Validate() error
}

// Type compliance.
var _ SCM = (*scm)(nil)
