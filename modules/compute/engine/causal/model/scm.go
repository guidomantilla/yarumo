package model

import (
	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/compute/math/graph"

	"github.com/guidomantilla/yarumo/compute/engine/causal"
)

type scm struct {
	g     *graph.Directed
	order []string
}

// NewSCM creates a new empty structural causal model.
func NewSCM() SCM {
	return &scm{
		g:     graph.NewDirected(),
		order: make([]string, 0),
	}
}

func (s *scm) AddVariable(name string, parents []string, equation EquationFn) error {
	cassert.NotNil(s, "scm is nil")

	if s.g.HasNode(name) {
		return causal.ErrCausal(causal.ErrDuplicateVariable)
	}

	if equation == nil {
		return causal.ErrCausal(causal.ErrNilEquation)
	}

	v := Variable{
		Name:     name,
		Parents:  parents,
		Equation: equation,
	}

	_ = s.g.AddNode(graph.Node{ID: name, Metadata: v})
	s.order = append(s.order, name)

	for _, p := range parents {
		if s.g.HasNode(p) {
			_ = s.g.AddEdge(graph.Edge{
				ID:   p + "->" + name,
				From: p,
				To:   name,
			})
		}
	}

	return nil
}

func (s *scm) Variable(name string) (Variable, bool) {
	cassert.NotNil(s, "scm is nil")

	gn, err := s.g.Node(name)
	if err != nil {
		return Variable{}, false
	}

	v, ok := gn.Metadata.(Variable)
	cassert.True(ok, "metadata is not a Variable")

	return v, true
}

func (s *scm) Variables() []string {
	cassert.NotNil(s, "scm is nil")

	sorted, err := graph.TopologicalSort(s.g)
	if err != nil {
		return s.order
	}

	return sorted
}

func (s *scm) Parents(name string) []string {
	cassert.NotNil(s, "scm is nil")

	gn, err := s.g.Node(name)
	if err != nil {
		return nil
	}

	v, ok := gn.Metadata.(Variable)
	cassert.True(ok, "metadata is not a Variable")

	return v.Parents
}

func (s *scm) Children(name string) []string {
	cassert.NotNil(s, "scm is nil")

	neighbors, err := s.g.Neighbors(name)
	if err != nil {
		return nil
	}

	return neighbors
}

func (s *scm) Validate() error {
	cassert.NotNil(s, "scm is nil")

	// Check all parents exist and ensure edges are present.
	for _, name := range s.order {
		gn, _ := s.g.Node(name)
		v, ok := gn.Metadata.(Variable)
		cassert.True(ok, "metadata is not a Variable")

		for _, p := range v.Parents {
			if !s.g.HasNode(p) {
				return causal.ErrCausal(causal.ErrParentNotFound)
			}

			edgeID := p + "->" + name

			if !s.g.HasEdge(edgeID) {
				_ = s.g.AddEdge(graph.Edge{
					ID:   edgeID,
					From: p,
					To:   name,
				})
			}
		}
	}

	// Check for cycles.
	if graph.HasCycle(s.g) {
		return causal.ErrCausal(causal.ErrCyclicModel)
	}

	return nil
}
