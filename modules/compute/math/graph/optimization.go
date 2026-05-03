package graph

import (
	"container/heap"
	"math"
	"slices"
)

// MinimumSpanningTree returns the edges forming an MST using Prim's algorithm.
// The graph must be connected.
func MinimumSpanningTree(g *Undirected) ([]string, error) {
	nodes := g.Nodes()
	if len(nodes) == 0 {
		return nil, nil
	}

	inMST := make(map[string]bool)
	var mstEdges []string

	pq := &edgePQ{}
	heap.Init(pq)

	addEdges := func(nodeID string) {
		inMST[nodeID] = true

		for _, eid := range g.adj[nodeID] {
			edge := g.edges[eid]
			other := edge.To

			if other == nodeID {
				other = edge.From
			}

			if !inMST[other] {
				heap.Push(pq, &edgePQItem{edgeID: eid, weight: edge.Weight, to: other})
			}
		}
	}

	addEdges(nodes[0].ID)

	for pq.Len() > 0 {
		raw := heap.Pop(pq)
		item, _ := raw.(*edgePQItem)

		if inMST[item.to] {
			continue
		}

		mstEdges = append(mstEdges, item.edgeID)
		addEdges(item.to)
	}

	if len(inMST) != len(nodes) {
		return nil, ErrGraph(ErrDisconnected)
	}

	slices.Sort(mstEdges)

	return mstEdges, nil
}

// MaxFlow computes the maximum flow from source to sink using Edmonds-Karp (BFS-based Ford-Fulkerson).
// Edge weights are treated as capacities.
func MaxFlow(g *Directed, source, sink string) (float64, error) {
	if !g.HasNode(source) || !g.HasNode(sink) {
		return 0, ErrGraph(ErrNodeNotFound)
	}

	capacity := make(map[string]map[string]float64)
	nodeIDs := make([]string, 0, g.NodeCount())

	for _, n := range g.Nodes() {
		nodeIDs = append(nodeIDs, n.ID)
		capacity[n.ID] = make(map[string]float64)
	}

	for _, e := range g.Edges() {
		capacity[e.From][e.To] += e.Weight
	}

	totalFlow := 0.0

	for {
		parent := bfsAugmentingPath(nodeIDs, capacity, source, sink)
		if parent == nil {
			break
		}

		pathFlow := math.Inf(1)

		for v := sink; v != source; v = parent[v] {
			u := parent[v]
			if capacity[u][v] < pathFlow {
				pathFlow = capacity[u][v]
			}
		}

		for v := sink; v != source; v = parent[v] {
			u := parent[v]
			capacity[u][v] -= pathFlow
			capacity[v][u] += pathFlow
		}

		totalFlow += pathFlow
	}

	return totalFlow, nil
}

// MinCut returns the minimum cut edges separating source from sink.
func MinCut(g *Directed, source, sink string) ([]string, error) {
	if !g.HasNode(source) || !g.HasNode(sink) {
		return nil, ErrGraph(ErrNodeNotFound)
	}

	capacity := make(map[string]map[string]float64)
	nodeIDs := make([]string, 0, g.NodeCount())

	for _, n := range g.Nodes() {
		nodeIDs = append(nodeIDs, n.ID)
		capacity[n.ID] = make(map[string]float64)
	}

	for _, e := range g.Edges() {
		capacity[e.From][e.To] += e.Weight
	}

	for {
		parent := bfsAugmentingPath(nodeIDs, capacity, source, sink)
		if parent == nil {
			break
		}

		pathFlow := math.Inf(1)

		for v := sink; v != source; v = parent[v] {
			u := parent[v]

			if capacity[u][v] < pathFlow {
				pathFlow = capacity[u][v]
			}
		}

		for v := sink; v != source; v = parent[v] {
			u := parent[v]
			capacity[u][v] -= pathFlow
			capacity[v][u] += pathFlow
		}
	}

	reachable := make(map[string]bool)
	queue := []string{source}
	reachable[source] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, nid := range nodeIDs {
			if !reachable[nid] && capacity[curr][nid] > 0 {
				reachable[nid] = true
				queue = append(queue, nid)
			}
		}
	}

	var cutEdges []string

	for _, e := range g.Edges() {
		if reachable[e.From] && !reachable[e.To] {
			cutEdges = append(cutEdges, e.ID)
		}
	}

	slices.Sort(cutEdges)

	return cutEdges, nil
}

// BipartiteMatching returns a maximum matching in a bipartite graph using Hopcroft-Karp.
func BipartiteMatching(g *Bipartite) []string {
	matchLeft := make(map[string]string)
	matchRight := make(map[string]string)
	left := g.Left()

	for {
		dist := make(map[string]int)
		queue := make([]string, 0)

		for _, u := range left {
			if matchLeft[u] == "" {
				dist[u] = 0
				queue = append(queue, u)
			}
		}

		found := false

		for len(queue) > 0 {
			u := queue[0]
			queue = queue[1:]

			neighbors, _ := g.Neighbors(u)

			for _, v := range neighbors {
				w := matchRight[v]

				if w == "" {
					found = true
				} else if dist[w] == 0 && w != "" {
					dist[w] = dist[u] + 1
					queue = append(queue, w)
				}
			}
		}

		if !found {
			break
		}

		for _, u := range left {
			if matchLeft[u] == "" {
				hopcroftDFS(g, u, matchLeft, matchRight, dist)
			}
		}
	}

	var result []string

	for _, u := range left {
		if matchLeft[u] != "" {
			edges := g.Edges()

			for _, e := range edges {
				if (e.From == u && e.To == matchLeft[u]) || (e.From == matchLeft[u] && e.To == u) {
					result = append(result, e.ID)

					break
				}
			}
		}
	}

	slices.Sort(result)

	return result
}

func hopcroftDFS(g *Bipartite, u string, matchLeft, matchRight map[string]string, dist map[string]int) bool {
	neighbors, _ := g.Neighbors(u)

	for _, v := range neighbors {
		w := matchRight[v]

		if w == "" || (dist[w] == dist[u]+1 && hopcroftDFS(g, w, matchLeft, matchRight, dist)) {
			matchLeft[u] = v
			matchRight[v] = u

			return true
		}
	}

	dist[u] = math.MaxInt

	return false
}

func bfsAugmentingPath(nodeIDs []string, capacity map[string]map[string]float64, source, sink string) map[string]string {
	parent := make(map[string]string)
	visited := make(map[string]bool)
	visited[source] = true
	queue := []string{source}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, nid := range nodeIDs {
			if !visited[nid] && capacity[curr][nid] > 0 {
				visited[nid] = true
				parent[nid] = curr

				if nid == sink {
					return parent
				}

				queue = append(queue, nid)
			}
		}
	}

	return nil
}

// Edge priority queue for MST.
type edgePQItem struct {
	edgeID string
	weight float64
	to     string
	index  int
}

type edgePQ []*edgePQItem

func (pq *edgePQ) Len() int { return len(*pq) }

func (pq *edgePQ) Less(i, j int) bool { return (*pq)[i].weight < (*pq)[j].weight }

func (pq *edgePQ) Swap(i, j int) {
	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
	(*pq)[i].index = i
	(*pq)[j].index = j
}

func (pq *edgePQ) Push(x any) {
	item, _ := x.(*edgePQItem)
	item.index = len(*pq)
	*pq = append(*pq, item)
}

func (pq *edgePQ) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[:n-1]

	return item
}
