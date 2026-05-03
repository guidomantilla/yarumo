package network

import (
	"errors"
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/stats"

	"github.com/guidomantilla/yarumo/compute/engine/bayesian"
)

func makeRainNetwork() Network {
	bn := NewNetwork()

	// Rain node (no parents).
	rainCPT := bayesian.NewCPT("Rain", nil)
	rainCPT.Set(stats.Assignment{}, stats.Distribution{"true": 0.2, "false": 0.8})

	bn.AddNode(Node{
		Variable: "Rain",
		Parents:  nil,
		CPT:      rainCPT,
		Outcomes: []stats.Outcome{"true", "false"},
	})

	// Sprinkler | Rain.
	sprinklerCPT := bayesian.NewCPT("Sprinkler", []stats.Var{"Rain"})
	sprinklerCPT.Set(stats.Assignment{"Rain": "true"}, stats.Distribution{"true": 0.01, "false": 0.99})
	sprinklerCPT.Set(stats.Assignment{"Rain": "false"}, stats.Distribution{"true": 0.4, "false": 0.6})

	bn.AddNode(Node{
		Variable: "Sprinkler",
		Parents:  []stats.Var{"Rain"},
		CPT:      sprinklerCPT,
		Outcomes: []stats.Outcome{"true", "false"},
	})

	// WetGrass | Rain, Sprinkler.
	wetCPT := bayesian.NewCPT("WetGrass", []stats.Var{"Rain", "Sprinkler"})
	wetCPT.Set(stats.Assignment{"Rain": "true", "Sprinkler": "true"}, stats.Distribution{"true": 0.99, "false": 0.01})
	wetCPT.Set(stats.Assignment{"Rain": "true", "Sprinkler": "false"}, stats.Distribution{"true": 0.8, "false": 0.2})
	wetCPT.Set(stats.Assignment{"Rain": "false", "Sprinkler": "true"}, stats.Distribution{"true": 0.9, "false": 0.1})
	wetCPT.Set(stats.Assignment{"Rain": "false", "Sprinkler": "false"}, stats.Distribution{"true": 0.0, "false": 1.0})

	bn.AddNode(Node{
		Variable: "WetGrass",
		Parents:  []stats.Var{"Rain", "Sprinkler"},
		CPT:      wetCPT,
		Outcomes: []stats.Outcome{"true", "false"},
	})

	return bn
}

func TestNewNetwork(t *testing.T) {
	t.Parallel()

	bn := NewNetwork()

	if len(bn.Nodes()) != 0 {
		t.Fatalf("expected empty network, got %d nodes", len(bn.Nodes()))
	}
}

func TestNetwork_AddNode(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()

	if len(bn.Nodes()) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(bn.Nodes()))
	}
}

func TestNetwork_Node(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()

	node, ok := bn.Node("Rain")
	if !ok {
		t.Fatal("expected Rain node")
	}

	if node.Variable != "Rain" {
		t.Fatalf("expected Rain, got %s", string(node.Variable))
	}
}

func TestNetwork_Node_notFound(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()

	_, ok := bn.Node("Unknown")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestNetwork_Parents(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()

	parents := bn.Parents("WetGrass")
	if len(parents) != 2 {
		t.Fatalf("expected 2 parents, got %d", len(parents))
	}
}

func TestNetwork_Parents_notFound(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()

	parents := bn.Parents("Unknown")
	if parents != nil {
		t.Fatalf("expected nil, got %v", parents)
	}
}

func TestNetwork_Children(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()

	children := bn.Children("Rain")
	if len(children) != 2 {
		t.Fatalf("expected 2 children (Sprinkler, WetGrass), got %d", len(children))
	}
}

func TestNetwork_Children_leaf(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()

	children := bn.Children("WetGrass")
	if len(children) != 0 {
		t.Fatalf("expected 0 children, got %d", len(children))
	}
}

func TestNetwork_TopologicalOrder(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()

	order := bn.TopologicalOrder()
	if len(order) != 3 {
		t.Fatalf("expected 3 variables, got %d", len(order))
	}

	// Rain should come first (no parents).
	if order[0] != "Rain" {
		t.Fatalf("expected Rain first, got %s", string(order[0]))
	}
}

func TestNetwork_Validate_valid(t *testing.T) {
	t.Parallel()

	bn := makeRainNetwork()

	err := bn.Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNetwork_Validate_invalidCPT(t *testing.T) {
	t.Parallel()

	bn := NewNetwork()

	badCPT := bayesian.NewCPT("X", nil)
	badCPT.Set(stats.Assignment{}, stats.Distribution{"a": 0.3, "b": 0.3}) // not normalized

	bn.AddNode(Node{
		Variable: "X",
		CPT:      badCPT,
		Outcomes: []stats.Outcome{"a", "b"},
	})

	err := bn.Validate()
	if !errors.Is(err, stats.ErrNotNormalized) {
		t.Fatalf("expected ErrNotNormalized wrapped, got %v", err)
	}
}

func TestNetwork_Validate_cyclic(t *testing.T) {
	t.Parallel()

	bn := NewNetwork()

	// A -> B -> A (cycle).
	aCPT := bayesian.NewCPT("A", []stats.Var{"B"})
	aCPT.Set(stats.Assignment{"B": "t"}, stats.Distribution{"t": 0.5, "f": 0.5})
	aCPT.Set(stats.Assignment{"B": "f"}, stats.Distribution{"t": 0.5, "f": 0.5})

	bCPT := bayesian.NewCPT("B", []stats.Var{"A"})
	bCPT.Set(stats.Assignment{"A": "t"}, stats.Distribution{"t": 0.5, "f": 0.5})
	bCPT.Set(stats.Assignment{"A": "f"}, stats.Distribution{"t": 0.5, "f": 0.5})

	bn.AddNode(Node{Variable: "A", Parents: []stats.Var{"B"}, CPT: aCPT, Outcomes: []stats.Outcome{"t", "f"}})
	bn.AddNode(Node{Variable: "B", Parents: []stats.Var{"A"}, CPT: bCPT, Outcomes: []stats.Outcome{"t", "f"}})

	err := bn.Validate()
	if !errors.Is(err, bayesian.ErrCyclicNetwork) {
		t.Fatalf("expected ErrCyclicNetwork, got %v", err)
	}
}

func TestNode_struct(t *testing.T) {
	t.Parallel()

	n := Node{
		Variable: "X",
		Parents:  []stats.Var{"Y"},
		Outcomes: []stats.Outcome{"a", "b"},
	}

	if n.Variable != "X" {
		t.Fatalf("expected X, got %s", string(n.Variable))
	}

	if len(n.Parents) != 1 {
		t.Fatalf("expected 1 parent, got %d", len(n.Parents))
	}
}
