package graph

// IsDAG reports whether a directed graph has no cycles.
func IsDAG(g *Directed) bool {
	return !hasCycleDirected(g)
}

// IsTree reports whether a directed graph is a rooted tree.
func IsTree(g *Directed) bool {
	r := roots(g)
	if len(r) != 1 {
		return false
	}

	for _, n := range g.Nodes() {
		if n.ID == r[0] {
			continue
		}

		if len(g.in[n.ID]) != 1 {
			return false
		}
	}

	return !hasCycleDirected(g)
}

// IsBipartite reports whether an undirected graph is bipartite.
func IsBipartite(g *Undirected) bool {
	_, _, ok := checkBipartite(g)
	return ok
}

// InDegree returns the in-degree of a node in a directed graph.
func InDegree(g *Directed, id string) (int, error) {
	if !g.HasNode(id) {
		return 0, ErrGraph(ErrNodeNotFound)
	}

	return len(g.in[id]), nil
}

// OutDegree returns the out-degree of a node in a directed graph.
func OutDegree(g *Directed, id string) (int, error) {
	if !g.HasNode(id) {
		return 0, ErrGraph(ErrNodeNotFound)
	}

	return len(g.out[id]), nil
}
