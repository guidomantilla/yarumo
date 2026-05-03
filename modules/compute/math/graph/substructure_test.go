package graph

import (
	"errors"
	"testing"
)

func TestFindCliques_triangle(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "A", To: "C"})

	cliques := FindCliques(g)
	if len(cliques) != 1 {
		t.Fatalf("expected 1 clique, got %d: %v", len(cliques), cliques)
	}
	if len(cliques[0]) != 3 {
		t.Fatalf("expected clique of size 3, got %d", len(cliques[0]))
	}
}

func TestFindCliques_no_edges(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})

	cliques := FindCliques(g)
	if len(cliques) != 2 {
		t.Fatalf("expected 2 singleton cliques, got %d", len(cliques))
	}
}

func TestFindCliques_two_connected_triangles(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "A", To: "C"})
	_ = g.AddEdge(Edge{ID: "e4", From: "C", To: "D"})
	_ = g.AddEdge(Edge{ID: "e5", From: "B", To: "D"})

	cliques := FindCliques(g)
	found3 := 0

	for _, c := range cliques {
		if len(c) == 3 {
			found3++
		}
	}

	if found3 != 2 {
		t.Fatalf("expected 2 triangles, got %d cliques: %v", found3, cliques)
	}
}

func TestFindCliques_empty(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	cliques := FindCliques(g)
	// Bron-Kerbosch with empty graph produces one empty clique
	if len(cliques) != 1 || len(cliques[0]) != 0 {
		t.Fatalf("expected 1 empty clique, got %d: %v", len(cliques), cliques)
	}
}

func TestFindCliques_four_node_path(t *testing.T) {
	t.Parallel()

	// A-B-C-D path graph forces X set to be non-empty with P empty,
	// exercising the pivot from X branch.
	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "C", To: "D"})

	cliques := FindCliques(g)
	if len(cliques) != 3 {
		t.Fatalf("expected 3 cliques (edges), got %d: %v", len(cliques), cliques)
	}
	for _, c := range cliques {
		if len(c) != 2 {
			t.Fatalf("expected all cliques of size 2, got %d: %v", len(c), c)
		}
	}
}

func TestFindCliques_diamond(t *testing.T) {
	t.Parallel()

	// Diamond: A-B, A-C, B-C, B-D, C-D → two triangles sharing edge B-C.
	// More complex recursion pattern for Bron-Kerbosch.
	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e4", From: "B", To: "D"})
	_ = g.AddEdge(Edge{ID: "e5", From: "C", To: "D"})

	cliques := FindCliques(g)
	found3 := 0
	for _, c := range cliques {
		if len(c) == 3 {
			found3++
		}
	}
	if found3 != 2 {
		t.Fatalf("expected 2 triangles, got %d in %v", found3, cliques)
	}
}

func TestEulerianPath_circuit(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "C", To: "A"})

	path, err := EulerianPath(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(path) != 4 {
		t.Fatalf("expected 4 nodes in circuit, got %d: %v", len(path), path)
	}
}

func TestEulerianPath_path(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})

	path, err := EulerianPath(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(path) != 3 {
		t.Fatalf("expected 3 nodes, got %d: %v", len(path), path)
	}
}

func TestEulerianPath_impossible(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "A", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "A", To: "D"})

	_, err := EulerianPath(g)
	if !errors.Is(err, ErrNoPath) {
		t.Fatalf("expected ErrNoPath, got %v", err)
	}
}

func TestEulerianPath_empty(t *testing.T) {
	t.Parallel()

	g := NewUndirected()
	path, err := EulerianPath(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != nil {
		t.Fatalf("expected nil, got %v", path)
	}
}

func TestEulerianPath_square_circuit(t *testing.T) {
	t.Parallel()

	// Square: A-B-C-D-A. In undirected IncidentEdges, edge orientation
	// may not match traversal direction, exercising edge.To == curr reversal.
	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "C"})
	_ = g.AddEdge(Edge{ID: "e3", From: "C", To: "D"})
	_ = g.AddEdge(Edge{ID: "e4", From: "D", To: "A"})

	path, err := EulerianPath(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(path) != 5 {
		t.Fatalf("expected 5 nodes in circuit, got %d: %v", len(path), path)
	}
	if path[0] != path[len(path)-1] {
		t.Fatal("expected circuit (first == last)")
	}
}

func TestEulerianPath_disconnected(t *testing.T) {
	t.Parallel()

	// 4 odd-degree nodes → no Eulerian path exists
	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "C", To: "D"})

	_, err := EulerianPath(g)
	if err == nil {
		t.Fatal("expected error for disconnected graph with 4 odd-degree nodes")
	}
}

func TestEulerianPath_disconnected_even_degree(t *testing.T) {
	t.Parallel()

	// Even degrees but disconnected → should detect disconnect
	g := NewUndirected()
	_ = g.AddNode(Node{ID: "A"})
	_ = g.AddNode(Node{ID: "B"})
	_ = g.AddNode(Node{ID: "C"})
	_ = g.AddNode(Node{ID: "D"})
	_ = g.AddEdge(Edge{ID: "e1", From: "A", To: "B"})
	_ = g.AddEdge(Edge{ID: "e2", From: "B", To: "A"})
	_ = g.AddEdge(Edge{ID: "e3", From: "C", To: "D"})
	_ = g.AddEdge(Edge{ID: "e4", From: "D", To: "C"})

	_, err := EulerianPath(g)
	if !errors.Is(err, ErrDisconnected) {
		t.Fatalf("expected ErrDisconnected, got %v", err)
	}
}
