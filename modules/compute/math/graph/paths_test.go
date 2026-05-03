package graph

import (
	"errors"
	"math"
	"testing"
)

func TestShortestPathDijkstra(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 1})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C", Weight: 2})
	_ = g.AddEdge(Edge{ID: "e3", From: "A", To: "C", Weight: 5})
	_ = g.AddEdge(Edge{ID: "e4", From: "C", To: "D", Weight: 1})

	p, err := ShortestPathDijkstra(g, "A", "D")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Weight != 4.0 {
		t.Fatalf("expected weight 4.0, got %f", p.Weight)
	}
	if len(p.Nodes) != 4 || p.Nodes[0] != "A" || p.Nodes[3] != "D" {
		t.Fatalf("unexpected path nodes: %v", p.Nodes)
	}
}

func TestShortestPathDijkstra_no_path(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})

	_, err := ShortestPathDijkstra(g, "A", "B")
	if !errors.Is(err, ErrNoPath) {
		t.Fatalf("expected ErrNoPath, got %v", err)
	}
}

func TestShortestPathDijkstra_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := ShortestPathDijkstra(g, "A", "B")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestShortestPathDijkstra_same_node(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})

	p, err := ShortestPathDijkstra(g, "A", "A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Weight != 0 {
		t.Fatalf("expected weight 0, got %f", p.Weight)
	}
}

func TestShortestPathBellmanFord(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 1})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C", Weight: 2})
	_ = g.AddEdge(Edge{ID: "e3", From: "A", To: "C", Weight: 5})

	p, err := ShortestPathBellmanFord(g, "A", "C")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Weight != 3.0 {
		t.Fatalf("expected weight 3.0, got %f", p.Weight)
	}
}

func TestShortestPathBellmanFord_negative_cycle(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 1})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C", Weight: -3})
	_ = g.AddEdge(Edge{ID: "e3", From: "C", To: "A", Weight: 1})

	_, err := ShortestPathBellmanFord(g, "A", "C")
	if !errors.Is(err, ErrNegativeCycle) {
		t.Fatalf("expected ErrNegativeCycle, got %v", err)
	}
}

func TestShortestPathBellmanFord_no_path(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})

	_, err := ShortestPathBellmanFord(g, "A", "B")
	if !errors.Is(err, ErrNoPath) {
		t.Fatalf("expected ErrNoPath, got %v", err)
	}
}

func TestShortestPathBellmanFord_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := ShortestPathBellmanFord(g, "X", "Y")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestAllPairsShortestPath(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 1})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C", Weight: 2})
	_ = g.AddEdge(Edge{ID: "e3", From: "A", To: "C", Weight: 5})

	dist := AllPairsShortestPath(g)
	if dist["A"]["C"] != 3.0 {
		t.Fatalf("expected A->C distance 3.0, got %f", dist["A"]["C"])
	}
	if dist["A"]["A"] != 0 {
		t.Fatalf("expected A->A distance 0, got %f", dist["A"]["A"])
	}
	if !math.IsInf(dist["C"]["A"], 1) {
		t.Fatalf("expected C->A distance Inf, got %f", dist["C"]["A"])
	}
}

func TestAllPaths(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 1})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C", Weight: 2})
	_ = g.AddEdge(Edge{ID: "e3", From: "A", To: "C", Weight: 5})

	paths, err := AllPaths(g, "A", "C")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}
	if paths[0].Weight != 3.0 {
		t.Fatalf("expected shortest path weight 3.0, got %f", paths[0].Weight)
	}
	if paths[1].Weight != 5.0 {
		t.Fatalf("expected second path weight 5.0, got %f", paths[1].Weight)
	}
}

func TestAllPaths_no_path(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})

	paths, err := AllPaths(g, "A", "B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 0 {
		t.Fatalf("expected 0 paths, got %d", len(paths))
	}
}

func TestAllPaths_not_found(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_, err := AllPaths(g, "X", "Y")
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestAllPaths_same_node(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})

	paths, _ := AllPaths(g, "A", "A")
	if len(paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(paths))
	}
	if paths[0].Weight != 0 {
		t.Fatalf("expected weight 0, got %f", paths[0].Weight)
	}
}

func TestAllPairsShortestPath_empty(t *testing.T) {
	t.Parallel()

	g := NewDirected()
	dist := AllPairsShortestPath(g)
	if len(dist) != 0 {
		t.Fatalf("expected empty, got %v", dist)
	}
}

func TestShortestPathDijkstra_stale_queue_entry(t *testing.T) {
	t.Parallel()

	// A→B(1), A→C(2), B→D(1), C→D(1), D→E(10)
	// D pushed with priority 2 (via B) and 3 (via C).
	// D(2) pops first, pushes E(12). Queue: D(3), E(12).
	// D(3) pops next and should be skipped as stale (3 > dist[D]=2).
	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddNode(Node{ID: "E"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 1})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "C", Weight: 2})
	_ = g.AddEdge(Edge{ID: "e3", From: "B", To: "D", Weight: 1})
	_ = g.AddEdge(Edge{ID: "e4", From: "C", To: "D", Weight: 1})
	_ = g.AddEdge(Edge{ID: "e5", From: "D", To: "E", Weight: 10})

	p, err := ShortestPathDijkstra(g, "A", "E")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Weight != 12.0 {
		t.Fatalf("expected weight 12.0, got %f", p.Weight)
	}
}

func TestShortestPathBellmanFord_disconnected(t *testing.T) {
	t.Parallel()

	// C is unreachable from A, exercises the IsInf skip.
	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 1})
	_ = g.AddEdge(Edge{ID: "e2", From: "C", To: "D", Weight: 1})

	_, err := ShortestPathBellmanFord(g, "A", "D")
	if err == nil {
		t.Fatal("expected error for unreachable destination")
	}
}

func TestAllPaths_equal_weight(t *testing.T) {
	t.Parallel()

	// Two paths with equal weight to exercise the return 0 branch in sort.
	g := NewDirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B", Weight: 2})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "D", Weight: 3})
	_ = g.AddEdge(Edge{ID: "e3", From: "A", To: "C", Weight: 1})
	_ = g.AddEdge(Edge{ID: "e4", From: "C", To: "D", Weight: 4})

	paths, err := AllPaths(g, "A", "D")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}
	if paths[0].Weight != 5.0 || paths[1].Weight != 5.0 {
		t.Fatalf("expected both paths weight 5.0, got %f and %f", paths[0].Weight, paths[1].Weight)
	}
}
