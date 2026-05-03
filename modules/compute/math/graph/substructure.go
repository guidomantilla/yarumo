package graph

import (
	"slices"
)

// FindCliques returns all maximal cliques using Bron-Kerbosch algorithm.
// Each clique is a sorted slice of node IDs.
func FindCliques(g *Undirected) [][]string {
	var result [][]string
	nodes := g.Nodes()
	nodeIDs := make([]string, len(nodes))

	for i, n := range nodes {
		nodeIDs[i] = n.ID
	}

	var bronKerbosch func(r, p, x []string)
	bronKerbosch = func(r, p, x []string) {
		if len(p) == 0 && len(x) == 0 {
			clique := make([]string, len(r))
			copy(clique, r)
			slices.Sort(clique)
			result = append(result, clique)

			return
		}

		pivot := ""
		if len(p) > 0 {
			pivot = p[0]
		} else if len(x) > 0 {
			pivot = x[0]
		}

		pivotNeighbors, _ := g.Neighbors(pivot)
		pivotSet := make(map[string]bool, len(pivotNeighbors))

		for _, nb := range pivotNeighbors {
			pivotSet[nb] = true
		}

		candidates := make([]string, 0, len(p))

		for _, v := range p {
			if !pivotSet[v] {
				candidates = append(candidates, v)
			}
		}

		for _, v := range candidates {
			neighbors, _ := g.Neighbors(v)
			nbSet := make(map[string]bool, len(neighbors))

			for _, nb := range neighbors {
				nbSet[nb] = true
			}

			newR := append(r, v) //nolint:gocritic
			newP := intersectSorted(p, nbSet)
			newX := intersectSorted(x, nbSet)

			bronKerbosch(newR, newP, newX)

			p = removeItem(p, v)
			x = append(x, v)
		}
	}

	bronKerbosch(nil, nodeIDs, nil)

	return result
}

// EulerianPath finds an Eulerian path in an undirected graph if one exists.
// Returns the sequence of node IDs forming the path.
func EulerianPath(g *Undirected) ([]string, error) {
	if g.EdgeCount() == 0 {
		return nil, nil
	}

	oddDegreeNodes := 0
	startNode := ""

	for _, n := range g.Nodes() {
		deg, _ := g.Degree(n.ID)

		if deg%2 != 0 {
			oddDegreeNodes++

			if startNode == "" {
				startNode = n.ID
			}
		}
	}

	if oddDegreeNodes != 0 && oddDegreeNodes != 2 {
		return nil, ErrGraph(ErrNoPath)
	}

	if startNode == "" {
		startNode = g.Nodes()[0].ID
	}

	usedEdges := make(map[string]bool)
	adjCopy := make(map[string][]string)

	for _, n := range g.Nodes() {
		edges, _ := g.IncidentEdges(n.ID)
		adjCopy[n.ID] = make([]string, len(edges))
		copy(adjCopy[n.ID], edges)
	}

	var circuit []string
	stack := []string{startNode}

	for len(stack) > 0 {
		curr := stack[len(stack)-1]

		foundEdge := false

		for len(adjCopy[curr]) > 0 {
			eid := adjCopy[curr][0]
			adjCopy[curr] = adjCopy[curr][1:]

			if usedEdges[eid] {
				continue
			}

			usedEdges[eid] = true
			edge, _ := g.Edge(eid)
			next := edge.To

			if next == curr {
				next = edge.From
			}

			stack = append(stack, next)
			foundEdge = true

			break
		}

		if !foundEdge {
			circuit = append(circuit, curr)
			stack = stack[:len(stack)-1]
		}
	}

	if len(usedEdges) != g.EdgeCount() {
		return nil, ErrGraph(ErrDisconnected)
	}

	slices.Reverse(circuit)

	return circuit, nil
}

func intersectSorted(s []string, set map[string]bool) []string {
	var result []string

	for _, v := range s {
		if set[v] {
			result = append(result, v)
		}
	}

	return result
}

func removeItem(s []string, item string) []string {
	var result []string

	for _, v := range s {
		if v != item {
			result = append(result, v)
		}
	}

	return result
}
