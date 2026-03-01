package network

import (
	"errors"
	"testing"

	"github.com/guidomantilla/yarumo/maths/probability"

	"github.com/guidomantilla/yarumo/inference/bayesian"
)

func makeRainNetwork() Network {
	bn := NewNetwork()

	// Rain node (no parents).
	rainCPT := probability.NewCPT("Rain", nil)
	rainCPT.Set(probability.Assignment{}, probability.Distribution{"true": 0.2, "false": 0.8})

	bn.AddNode(Node{
		Variable: "Rain",
		Parents:  nil,
		CPT:      rainCPT,
		Outcomes: []probability.Outcome{"true", "false"},
	})

	// Sprinkler | Rain.
	sprinklerCPT := probability.NewCPT("Sprinkler", []probability.Var{"Rain"})
	sprinklerCPT.Set(probability.Assignment{"Rain": "true"}, probability.Distribution{"true": 0.01, "false": 0.99})
	sprinklerCPT.Set(probability.Assignment{"Rain": "false"}, probability.Distribution{"true": 0.4, "false": 0.6})

	bn.AddNode(Node{
		Variable: "Sprinkler",
		Parents:  []probability.Var{"Rain"},
		CPT:      sprinklerCPT,
		Outcomes: []probability.Outcome{"true", "false"},
	})

	// WetGrass | Rain, Sprinkler.
	wetCPT := probability.NewCPT("WetGrass", []probability.Var{"Rain", "Sprinkler"})
	wetCPT.Set(probability.Assignment{"Rain": "true", "Sprinkler": "true"}, probability.Distribution{"true": 0.99, "false": 0.01})
	wetCPT.Set(probability.Assignment{"Rain": "true", "Sprinkler": "false"}, probability.Distribution{"true": 0.8, "false": 0.2})
	wetCPT.Set(probability.Assignment{"Rain": "false", "Sprinkler": "true"}, probability.Distribution{"true": 0.9, "false": 0.1})
	wetCPT.Set(probability.Assignment{"Rain": "false", "Sprinkler": "false"}, probability.Distribution{"true": 0.0, "false": 1.0})

	bn.AddNode(Node{
		Variable: "WetGrass",
		Parents:  []probability.Var{"Rain", "Sprinkler"},
		CPT:      wetCPT,
		Outcomes: []probability.Outcome{"true", "false"},
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

	badCPT := probability.NewCPT("X", nil)
	badCPT.Set(probability.Assignment{}, probability.Distribution{"a": 0.3, "b": 0.3}) // not normalized

	bn.AddNode(Node{
		Variable: "X",
		CPT:      badCPT,
		Outcomes: []probability.Outcome{"a", "b"},
	})

	err := bn.Validate()
	if !errors.Is(err, probability.ErrNotNormalized) {
		t.Fatalf("expected ErrNotNormalized wrapped, got %v", err)
	}
}

func TestNetwork_Validate_cyclic(t *testing.T) {
	t.Parallel()

	bn := NewNetwork()

	// A -> B -> A (cycle).
	aCPT := probability.NewCPT("A", []probability.Var{"B"})
	aCPT.Set(probability.Assignment{"B": "t"}, probability.Distribution{"t": 0.5, "f": 0.5})
	aCPT.Set(probability.Assignment{"B": "f"}, probability.Distribution{"t": 0.5, "f": 0.5})

	bCPT := probability.NewCPT("B", []probability.Var{"A"})
	bCPT.Set(probability.Assignment{"A": "t"}, probability.Distribution{"t": 0.5, "f": 0.5})
	bCPT.Set(probability.Assignment{"A": "f"}, probability.Distribution{"t": 0.5, "f": 0.5})

	bn.AddNode(Node{Variable: "A", Parents: []probability.Var{"B"}, CPT: aCPT, Outcomes: []probability.Outcome{"t", "f"}})
	bn.AddNode(Node{Variable: "B", Parents: []probability.Var{"A"}, CPT: bCPT, Outcomes: []probability.Outcome{"t", "f"}})

	err := bn.Validate()
	if !errors.Is(err, bayesian.ErrCyclicNetwork) {
		t.Fatalf("expected ErrCyclicNetwork, got %v", err)
	}
}

func TestNode_struct(t *testing.T) {
	t.Parallel()

	n := Node{
		Variable: "X",
		Parents:  []probability.Var{"Y"},
		Outcomes: []probability.Outcome{"a", "b"},
	}

	if n.Variable != "X" {
		t.Fatalf("expected X, got %s", string(n.Variable))
	}

	if len(n.Parents) != 1 {
		t.Fatalf("expected 1 parent, got %d", len(n.Parents))
	}
}
