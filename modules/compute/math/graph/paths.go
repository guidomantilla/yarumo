package graph

import (
	"container/heap"
	"math"
	"slices"
)

// ShortestPathDijkstra finds the shortest path from src to dst using Dijkstra's algorithm.
// Requires non-negative edge weights.
func ShortestPathDijkstra(g Graph, src, dst string) (Path, error) {
	if !g.HasNode(src) || !g.HasNode(dst) {
		return Path{}, ErrGraph(ErrNodeNotFound)
	}

	dist := make(map[string]float64)
	prev := make(map[string]string)
	prevEdge := make(map[string]string)

	for _, n := range g.Nodes() {
		dist[n.ID] = math.Inf(1)
	}

	dist[src] = 0

	pq := &priorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &pqItem{id: src, priority: 0})

	for pq.Len() > 0 {
		raw := heap.Pop(pq)
		item, _ := raw.(*pqItem)
		u := item.id

		if u == dst {
			break
		}

		if item.priority > dist[u] {
			continue
		}

		for _, e := range edgesFrom(g, u) {
			alt := dist[u] + e.Weight

			if alt < dist[e.To] {
				dist[e.To] = alt
				prev[e.To] = u
				prevEdge[e.To] = e.ID
				heap.Push(pq, &pqItem{id: e.To, priority: alt})
			}
		}
	}

	if math.IsInf(dist[dst], 1) {
		return Path{}, ErrGraph(ErrNoPath)
	}

	return buildPath(src, dst, prev, prevEdge, dist[dst]), nil
}

// ShortestPathBellmanFord finds the shortest path from src to dst using Bellman-Ford.
// Supports negative weights; detects negative cycles.
func ShortestPathBellmanFord(g Graph, src, dst string) (Path, error) {
	if !g.HasNode(src) || !g.HasNode(dst) {
		return Path{}, ErrGraph(ErrNodeNotFound)
	}

	nodes := g.Nodes()
	edges := allDirectionalEdges(g)
	dist := make(map[string]float64)
	prev := make(map[string]string)
	prevEdge := make(map[string]string)

	for _, n := range nodes {
		dist[n.ID] = math.Inf(1)
	}

	dist[src] = 0

	for range len(nodes) - 1 {
		for _, e := range edges {
			if math.IsInf(dist[e.From], 1) {
				continue
			}

			alt := dist[e.From] + e.Weight
			if alt < dist[e.To] {
				dist[e.To] = alt
				prev[e.To] = e.From
				prevEdge[e.To] = e.ID
			}
		}
	}

	for _, e := range edges {
		if math.IsInf(dist[e.From], 1) {
			continue
		}

		if dist[e.From]+e.Weight < dist[e.To] {
			return Path{}, ErrGraph(ErrNegativeCycle)
		}
	}

	if math.IsInf(dist[dst], 1) {
		return Path{}, ErrGraph(ErrNoPath)
	}

	return buildPath(src, dst, prev, prevEdge, dist[dst]), nil
}

// AllPairsShortestPath computes shortest distances between all pairs using Floyd-Warshall.
func AllPairsShortestPath(g Graph) map[string]map[string]float64 {
	nodes := g.Nodes()
	dist := make(map[string]map[string]float64, len(nodes))

	for _, u := range nodes {
		dist[u.ID] = make(map[string]float64, len(nodes))

		for _, v := range nodes {
			if u.ID == v.ID {
				dist[u.ID][v.ID] = 0
			} else {
				dist[u.ID][v.ID] = math.Inf(1)
			}
		}
	}

	for _, e := range allDirectionalEdges(g) {
		if e.Weight < dist[e.From][e.To] {
			dist[e.From][e.To] = e.Weight
		}
	}

	for _, k := range nodes {
		for _, i := range nodes {
			for _, j := range nodes {
				through := dist[i.ID][k.ID] + dist[k.ID][j.ID]
				if through < dist[i.ID][j.ID] {
					dist[i.ID][j.ID] = through
				}
			}
		}
	}

	return dist
}

// AllPaths returns all simple paths from src to dst, sorted by weight.
func AllPaths(g Graph, src, dst string) ([]Path, error) {
	if !g.HasNode(src) || !g.HasNode(dst) {
		return nil, ErrGraph(ErrNodeNotFound)
	}

	var results []Path
	visited := make(map[string]bool)
	var pathNodes []string
	var pathEdges []string
	var weight float64

	var dfs func(string)
	dfs = func(curr string) {
		visited[curr] = true
		pathNodes = append(pathNodes, curr)

		if curr == dst {
			p := Path{
				Nodes:  make([]string, len(pathNodes)),
				Edges:  make([]string, len(pathEdges)),
				Weight: weight,
			}

			copy(p.Nodes, pathNodes)
			copy(p.Edges, pathEdges)
			results = append(results, p)
		} else {
			for _, e := range edgesFrom(g, curr) {
				if visited[e.To] {
					continue
				}

				pathEdges = append(pathEdges, e.ID)
				weight += e.Weight
				dfs(e.To)
				pathEdges = pathEdges[:len(pathEdges)-1]
				weight -= e.Weight
			}
		}

		pathNodes = pathNodes[:len(pathNodes)-1]
		visited[curr] = false
	}

	dfs(src)

	slices.SortFunc(results, func(a, b Path) int {
		if a.Weight < b.Weight {
			return -1
		}

		if a.Weight > b.Weight {
			return 1
		}

		return 0
	})

	return results, nil
}

// allDirectionalEdges returns all edges with both directions for undirected graphs.
func allDirectionalEdges(g Graph) []Edge {
	edges := g.Edges()

	if g.IsDirected() {
		return edges
	}

	result := make([]Edge, 0, len(edges)*2)

	for _, e := range edges {
		result = append(result, e)

		if e.From != e.To {
			result = append(result, Edge{
				ID: e.ID, From: e.To, To: e.From,
				Weight: e.Weight, Label: e.Label, Metadata: e.Metadata,
			})
		}
	}

	return result
}

// edgesFrom returns edges usable from nodeID for path computation.
// For undirected graphs, edges where nodeID is the To endpoint are reversed.
func edgesFrom(g Graph, nodeID string) []Edge {
	var result []Edge

	for _, e := range g.Edges() {
		if e.From == nodeID {
			result = append(result, e)
		} else if !g.IsDirected() && e.To == nodeID {
			result = append(result, Edge{
				ID: e.ID, From: nodeID, To: e.From,
				Weight: e.Weight, Label: e.Label, Metadata: e.Metadata,
			})
		}
	}

	return result
}

func buildPath(src, dst string, prev, prevEdge map[string]string, weight float64) Path {
	var nodes []string
	var edges []string

	for curr := dst; curr != src; curr = prev[curr] {
		nodes = append(nodes, curr)
		edges = append(edges, prevEdge[curr])
	}

	nodes = append(nodes, src)
	slices.Reverse(nodes)
	slices.Reverse(edges)

	return Path{Nodes: nodes, Edges: edges, Weight: weight}
}

// Priority queue for Dijkstra.
type pqItem struct {
	id       string
	priority float64
	index    int
}

type priorityQueue []*pqItem

func (pq *priorityQueue) Len() int { return len(*pq) }

func (pq *priorityQueue) Less(i, j int) bool { return (*pq)[i].priority < (*pq)[j].priority }

func (pq *priorityQueue) Swap(i, j int) {
	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
	(*pq)[i].index = i
	(*pq)[j].index = j
}

func (pq *priorityQueue) Push(x any) {
	item, _ := x.(*pqItem)
	item.index = len(*pq)
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[:n-1]

	return item
}
