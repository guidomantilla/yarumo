package adapters

import (
	"errors"
	"testing"

	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

func TestAdaptBayesianNetwork(t *testing.T) {
	t.Parallel()

	t.Run("valid network", func(t *testing.T) {
		t.Parallel()

		config := &schema.BayesianConfig{
			Nodes: []schema.BayesianNodeDef{
				{
					Variable: "rain",
					Outcomes: []string{"yes", "no"},
					CPT: []schema.CPTRow{
						{
							Probabilities: map[string]float64{"yes": 0.3, "no": 0.7},
						},
					},
				},
				{
					Variable: "wet",
					Parents:  []string{"rain"},
					Outcomes: []string{"yes", "no"},
					CPT: []schema.CPTRow{
						{
							Given:         map[string]string{"rain": "yes"},
							Probabilities: map[string]float64{"yes": 0.9, "no": 0.1},
						},
						{
							Given:         map[string]string{"rain": "no"},
							Probabilities: map[string]float64{"yes": 0.2, "no": 0.8},
						},
					},
				},
			},
		}

		net, err := AdaptBayesianNetwork(config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		nodes := net.Nodes()
		if len(nodes) != 2 {
			t.Fatalf("expected 2 nodes, got %d", len(nodes))
		}
	})
}

func TestAdaptBayesianNetwork_ValidationError(t *testing.T) {
	t.Parallel()

	t.Run("invalid CPT probabilities", func(t *testing.T) {
		t.Parallel()

		config := &schema.BayesianConfig{
			Nodes: []schema.BayesianNodeDef{
				{
					Variable: "rain",
					Outcomes: []string{"yes", "no"},
					CPT: []schema.CPTRow{
						{Probabilities: map[string]float64{"yes": 0.5, "no": 0.8}},
					},
				},
			},
		}

		_, err := AdaptBayesianNetwork(config)
		if err == nil {
			t.Fatal("expected validation error")
		}

		if !errors.Is(err, ErrAdaptNetworkFailed) {
			t.Fatalf("expected ErrAdaptNetworkFailed, got: %v", err)
		}
	})

	t.Run("duplicate node", func(t *testing.T) {
		t.Parallel()

		config := &schema.BayesianConfig{
			Nodes: []schema.BayesianNodeDef{
				{
					Variable: "rain",
					Outcomes: []string{"yes", "no"},
					CPT: []schema.CPTRow{
						{Probabilities: map[string]float64{"yes": 0.3, "no": 0.7}},
					},
				},
				{
					Variable: "rain",
					Outcomes: []string{"yes", "no"},
					CPT: []schema.CPTRow{
						{Probabilities: map[string]float64{"yes": 0.4, "no": 0.6}},
					},
				},
			},
		}

		_, err := AdaptBayesianNetwork(config)
		if err == nil {
			t.Fatal("expected add-node error")
		}

		if !errors.Is(err, ErrAdaptNetworkFailed) {
			t.Fatalf("expected ErrAdaptNetworkFailed, got: %v", err)
		}
	})
}

func TestAdaptBayesianOpts(t *testing.T) {
	t.Parallel()

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		opts := AdaptBayesianOpts(&schema.BayesianConfig{})
		if len(opts) != 0 {
			t.Fatalf("expected 0 opts, got %d", len(opts))
		}
	})

	t.Run("variable elimination", func(t *testing.T) {
		t.Parallel()

		opts := AdaptBayesianOpts(&schema.BayesianConfig{Algorithm: "variable_elimination"})
		if len(opts) != 1 {
			t.Fatalf("expected 1 opt, got %d", len(opts))
		}
	})
}
