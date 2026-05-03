package graph

import (
	"errors"
	"testing"
)

func TestTopologicalSort(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "B", To: "D"})
	_ = g.AddEdge(Edge{ID: "e4", From: "C", To: "D"})

	order, err := TopologicalSort(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(order))
	}

	pos := make(map[string]int)
	for i, id := range order {
		pos[id] = i
	}

	if pos["A"] > pos["B"] || pos["A"] > pos["C"] {
		t.Fatal("A should come before B and C")
	}
	if pos["B"] > pos["D"] || pos["C"] > pos["D"] {
		t.Fatal("B and C should come before D")
	}
}

func TestTopologicalSort_cycle(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "A"})

	_, err := TopologicalSort(g)
	if !errors.Is(err, ErrNotDAG) {
		t.Fatalf("expected ErrNotDAG, got %v", err)
	}
}

func TestHasCycle_true(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "A"})

	if !HasCycle(g) {
		t.Fatal("expected cycle")
	}
}

func TestHasCycle_false(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	if HasCycle(g) {
		t.Fatal("expected no cycle")
	}
}

func TestConnectedComponents(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "C", To: "D"})

	cc := ConnectedComponents(g)
	if len(cc) != 2 {
		t.Fatalf("expected 2 components, got %d", len(cc))
	}
}

func TestConnectedComponents_single(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	cc := ConnectedComponents(g)
	if len(cc) != 1 {
		t.Fatalf("expected 1 component, got %d", len(cc))
	}
}

func TestStronglyConnectedComponents(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "C", To: "A"})
	_ = g.AddEdge(Edge{ID: "e4", From: "C", To: "D"})

	sccs := StronglyConnectedComponents(g)

	foundCycle := false
	foundSingleton := false

	for _, scc := range sccs {
		if len(scc) == 3 {
			foundCycle = true
		}

		if len(scc) == 1 && scc[0] == "D" {
			foundSingleton = true
		}
	}

	if !foundCycle {
		t.Fatal("expected SCC {A,B,C}")
	}

	if !foundSingleton {
		t.Fatal("expected singleton SCC {D}")
	}
}

func TestStronglyConnectedComponents_no_cycle(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	sccs := StronglyConnectedComponents(g)
	if len(sccs) != 2 {
		t.Fatalf("expected 2 SCCs, got %d", len(sccs))
	}
}

func TestReachable(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	reachable, err := Reachable(g, "A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reachable) != 2 {
		t.Fatalf("expected 2 reachable nodes, got %d", len(reachable))
	}
	if reachable[0] != "B" || reachable[1] != "C" {
		t.Fatalf("expected [B C], got %v", reachable)
	}
}

func TestReachable_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := Reachable(g, "X")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestAncestors(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	anc, err := Ancestors(g, "C")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(anc) != 2 || anc[0] != "A" || anc[1] != "B" {
		t.Fatalf("expected [A B], got %v", anc)
	}
}

func TestAncestors_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := Ancestors(g, "X")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestDescendants(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	desc, err := Descendants(g, "A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(desc) != 2 || desc[0] != "B" || desc[1] != "C" {
		t.Fatalf("expected [B C], got %v", desc)
	}
}

func TestBridges(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "C", To: "A"})
	_ = g.AddEdge(Edge{ID: "e4", From: "C", To: "D"})

	bridges := Bridges(g)
	if len(bridges) != 1 || bridges[0] != "e4" {
		t.Fatalf("expected [e4], got %v", bridges)
	}
}

func TestBridges_no_bridges(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "C", To: "A"})

	bridges := Bridges(g)
	if len(bridges) != 0 {
		t.Fatalf("expected no bridges, got %v", bridges)
	}
}

func TestArticulationPoints(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "C", To: "A"})
	_ = g.AddEdge(Edge{ID: "e4", From: "C", To: "D"})

	aps := ArticulationPoints(g)
	if len(aps) != 1 || aps[0] != "C" {
		t.Fatalf("expected [C], got %v", aps)
	}
}

func TestArticulationPoints_root_multiple_children(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "C"})

	aps := ArticulationPoints(g)
	if len(aps) != 1 || aps[0] != "A" {
		t.Fatalf("expected [A], got %v", aps)
	}
}

func TestDiameter(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	d := Diameter(g)
	if d != 2 {
		t.Fatalf("expected diameter 2, got %d", d)
	}
}

func TestDiameter_disconnected(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})

	d := Diameter(g)
	if d != -1 {
		t.Fatalf("expected -1 for disconnected, got %d", d)
	}
}

func TestDiameter_empty(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	d := Diameter(g)
	if d != 0 {
		t.Fatalf("expected 0, got %d", d)
	}
}

func TestDiameter_single_node(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})

	d := Diameter(g)
	if d != 0 {
		t.Fatalf("expected 0, got %d", d)
	}
}

func TestTopologicalSort_empty(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	order, err := TopologicalSort(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 0 {
		t.Fatalf("expected empty, got %v", order)
	}
}

func TestConnectedComponents_empty(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	cc := ConnectedComponents(g)
	if len(cc) != 0 {
		t.Fatalf("expected empty, got %v", cc)
	}
}

func TestReachable_isolated(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})

	reachable, _ := Reachable(g, "A")
	if len(reachable) != 0 {
		t.Fatalf("expected empty, got %v", reachable)
	}
}

func TestAncestors_root(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})

	anc, _ := Ancestors(g, "A")
	if len(anc) != 0 {
		t.Fatalf("expected no ancestors for root, got %v", anc)
	}
}

func TestBridges_disconnected(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "C", To: "D"})

	bridges := Bridges(g)
	if len(bridges) != 2 {
		t.Fatalf("expected 2 bridges, got %v", bridges)
	}
}
