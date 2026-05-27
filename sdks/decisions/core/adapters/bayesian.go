package adapters

import (
	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian"
	bengine "github.com/guidomantilla/yarumo/compute/engine/bayesian/engine"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/network"
	"github.com/guidomantilla/yarumo/compute/math/stats"

	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

// AdaptBayesianNode converts a single serialized node definition to a network node.
func AdaptBayesianNode(def schema.BayesianNodeDef) network.Node {
	parents := make([]stats.Var, len(def.Parents))
	for i, p := range def.Parents {
		parents[i] = stats.Var(p)
	}

	outcomes := make([]stats.Outcome, len(def.Outcomes))
	for i, o := range def.Outcomes {
		outcomes[i] = stats.Outcome(o)
	}

	cpt := bayesian.NewCPT(stats.Var(def.Variable), parents)

	for _, row := range def.CPT {
		given := make(stats.Assignment, len(row.Given))
		for k, v := range row.Given {
			given[stats.Var(k)] = stats.Outcome(v)
		}

		dist := make(stats.Distribution, len(row.Probabilities))
		for k, v := range row.Probabilities {
			dist[stats.Outcome(k)] = stats.Prob(v)
		}

		cpt.Set(given, dist)
	}

	return network.Node{
		Variable: stats.Var(def.Variable),
		Parents:  parents,
		CPT:      cpt,
		Outcomes: outcomes,
	}
}

// AdaptBayesianNetwork converts a serialized Bayesian config to an inference network.
func AdaptBayesianNetwork(config *schema.BayesianConfig) (network.Network, error) {
	cassert.NotNil(config, "config is nil")

	net := network.NewNetwork()

	for _, def := range config.Nodes {
		node := AdaptBayesianNode(def)

		err := net.AddNode(node)
		if err != nil {
			return nil, ErrAdaptNetwork(err)
		}
	}

	err := net.Validate()
	if err != nil {
		return nil, ErrAdaptNetwork(err)
	}

	return net, nil
}

// AdaptBayesianOpts converts serialized algorithm config to engine options.
func AdaptBayesianOpts(config *schema.BayesianConfig) []bengine.Option {
	cassert.NotNil(config, "config is nil")

	var opts []bengine.Option

	if config.Algorithm == "variable_elimination" {
		opts = append(opts, bengine.WithAlgorithm(bengine.VariableElimination))
	}

	return opts
}
