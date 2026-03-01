package engine

import (
	cassert "github.com/guidomantilla/yarumo/common/assert"
	"github.com/guidomantilla/yarumo/maths/probability"

	"github.com/guidomantilla/yarumo/inference/bayesian/evidence"
	"github.com/guidomantilla/yarumo/inference/bayesian/explain"
	"github.com/guidomantilla/yarumo/inference/bayesian/network"
)

type engine struct {
	options Options
}

// NewEngine creates a new Bayesian inference engine with the given options.
func NewEngine(opts ...Option) Engine {
	return &engine{
		options: NewOptions(opts...),
	}
}

func (e *engine) Query(net network.Network, ev evidence.EvidenceBase, query probability.Var) Result {
	cassert.NotNil(e, "engine is nil")

	observed := ev.Observed()
	trace := explain.NewTrace(query, observed)

	if e.options.algorithm == VariableElimination {
		return e.variableElimination(net, observed, query, trace)
	}

	return e.enumerate(net, observed, query, trace)
}
