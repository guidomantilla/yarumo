package graph

import (
	"math"
	"slices"
)

// TopologicalSort returns a topological ordering of a DAG's nodes.
func TopologicalSort(g *Directed) ([]string, error) {
	if hasCycleDirected(g) {
		return nil, ErrGraph(ErrNotDAG)
	}

	visited := make(map[string]bool)
	var order []string

	var visit func(string)
	visit = func(id string) {
		if visited[id] {
			return
		}

		visited[id] = true

		for _, eid := range g.out[id] {
			edge := g.edges[eid]
			visit(edge.To)
		}

		order = append(order, id)
	}

	for _, n := range g.Nodes() {
		visit(n.ID)
	}

	slices.Reverse(order)

	return order, nil
}

// HasCycle reports whether a directed graph contains a cycle.
func HasCycle(g *Directed) bool {
	return hasCycleDirected(g)
}

// ConnectedComponents returns the connected components of an undirected graph.
// Each component is a sorted slice of node IDs.
func ConnectedComponents(g *Undirected) [][]string {
	visited := make(map[string]bool)
	nodes := g.Nodes()
	components := make([][]string, 0, len(nodes))

	for _, n := range nodes {
		if visited[n.ID] {
			continue
		}

		var component []string

		_ = BFS(g, n.ID, func(id string) {
			component = append(component, id)
			visited[id] = true
		})

		slices.Sort(component)
		components = append(components, component)
	}

	return components
}

// StronglyConnectedComponents returns the SCCs of a directed graph using Tarjan's algorithm.
// Each SCC is a sorted slice of node IDs.
func StronglyConnectedComponents(g *Directed) [][]string {
	index := 0
	nodeIndex := make(map[string]int)
	lowlink := make(map[string]int)
	onStack := make(map[string]bool)
	var stack []string
	var result [][]string

	var strongConnect func(string)
	strongConnect = func(id string) {
		nodeIndex[id] = index
		lowlink[id] = index
		index++
		stack = append(stack, id)
		onStack[id] = true

		for _, eid := range g.out[id] {
			edge := g.edges[eid]
			w := edge.To

			if _, visited := nodeIndex[w]; !visited {
				strongConnect(w)
				lowlink[id] = min(lowlink[id], lowlink[w])
			} else if onStack[w] {
				lowlink[id] = min(lowlink[id], nodeIndex[w])
			}
		}

		if lowlink[id] == nodeIndex[id] {
			var component []string

			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				component = append(component, w)

				if w == id {
					break
				}
			}

			slices.Sort(component)
			result = append(result, component)
		}
	}

	for _, n := range g.Nodes() {
		if _, visited := nodeIndex[n.ID]; !visited {
			strongConnect(n.ID)
		}
	}

	return result
}

// Reachable returns all nodes reachable from the given node, sorted.
func Reachable(g Graph, id string) ([]string, error) {
	if !g.HasNode(id) {
		return nil, ErrGraph(ErrNodeNotFound)
	}

	var result []string

	_ = BFS(g, id, func(nid string) {
		if nid != id {
			result = append(result, nid)
		}
	})

	slices.Sort(result)

	return result, nil
}

// Ancestors returns all nodes that can reach the given node in a directed graph, sorted.
func Ancestors(g *Directed, id string) ([]string, error) {
	if !g.HasNode(id) {
		return nil, ErrGraph(ErrNodeNotFound)
	}

	visited := make(map[string]bool)
	var result []string

	var visit func(string)
	visit = func(nid string) {
		for _, eid := range g.in[nid] {
			edge := g.edges[eid]

			if !visited[edge.From] {
				visited[edge.From] = true
				result = append(result, edge.From)
				visit(edge.From)
			}
		}
	}

	visit(id)
	slices.Sort(result)

	return result, nil
}

// Descendants returns all nodes reachable from the given node in a directed graph, sorted.
func Descendants(g *Directed, id string) ([]string, error) {
	return Reachable(g, id)
}

// Bridges returns edges whose removal disconnects the undirected graph.
func Bridges(g *Undirected) []string {
	disc := make(map[string]int)
	low := make(map[string]int)
	timer := 0
	var bridges []string

	var dfs func(string, string)
	dfs = func(u, parentEdge string) {
		disc[u] = timer
		low[u] = timer
		timer++

		for _, eid := range g.adj[u] {
			if eid == parentEdge {
				continue
			}

			edge := g.edges[eid]
			v := edge.To

			if v == u {
				v = edge.From
			}

			if _, visited := disc[v]; !visited {
				dfs(v, eid)
				low[u] = min(low[u], low[v])

				if low[v] > disc[u] {
					bridges = append(bridges, eid)
				}
			} else {
				low[u] = min(low[u], disc[v])
			}
		}
	}

	for _, n := range g.Nodes() {
		if _, visited := disc[n.ID]; !visited {
			dfs(n.ID, "")
		}
	}

	slices.Sort(bridges)

	return bridges
}

// ArticulationPoints returns nodes whose removal disconnects the undirected graph, sorted.
func ArticulationPoints(g *Undirected) []string {
	disc := make(map[string]int)
	low := make(map[string]int)
	ap := make(map[string]bool)
	timer := 0

	var dfs func(string, string)
	dfs = func(u, parentEdge string) {
		disc[u] = timer
		low[u] = timer
		timer++
		childCount := 0

		for _, eid := range g.adj[u] {
			if eid == parentEdge {
				continue
			}

			edge := g.edges[eid]
			v := edge.To

			if v == u {
				v = edge.From
			}

			if _, visited := disc[v]; !visited {
				childCount++
				dfs(v, eid)
				low[u] = min(low[u], low[v])

				if parentEdge == "" && childCount > 1 {
					ap[u] = true
				}

				if parentEdge != "" && low[v] >= disc[u] {
					ap[u] = true
				}
			} else {
				low[u] = min(low[u], disc[v])
			}
		}
	}

	for _, n := range g.Nodes() {
		if _, visited := disc[n.ID]; !visited {
			dfs(n.ID, "")
		}
	}

	return sortedKeys(ap)
}

// Diameter returns the diameter of an undirected graph (longest shortest path).
// Returns -1 for disconnected graphs.
func Diameter(g *Undirected) int {
	nodes := g.Nodes()
	if len(nodes) == 0 {
		return 0
	}

	maxDist := 0

	for _, n := range nodes {
		dist := bfsDistances(g, n.ID)

		for _, d := range dist {
			if d == math.MaxInt {
				return -1
			}

			if d > maxDist {
				maxDist = d
			}
		}
	}

	return maxDist
}

func bfsDistances(g *Undirected, start string) map[string]int {
	dist := make(map[string]int)

	for _, n := range g.Nodes() {
		dist[n.ID] = math.MaxInt
	}

	dist[start] = 0
	queue := []string{start}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		neighbors, _ := g.Neighbors(curr)

		for _, nb := range neighbors {
			if dist[nb] == math.MaxInt {
				dist[nb] = dist[curr] + 1
				queue = append(queue, nb)
			}
		}
	}

	return dist
}
