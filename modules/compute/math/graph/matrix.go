package graph

import (
	"maps"
	"math"
	"slices"
)

// ToMatrix converts a graph to its adjacency matrix representation.
// Non-adjacent pairs are represented as +Inf.
func ToMatrix(g Graph) Matrix {
	nodes := g.Nodes()
	n := len(nodes)
	indices := make(map[string]int, n)

	for i, node := range nodes {
		indices[node.ID] = i
	}

	data := make([][]float64, n)

	for i := range n {
		data[i] = make([]float64, n)

		for j := range n {
			if i == j {
				data[i][j] = 0
			} else {
				data[i][j] = math.Inf(1)
			}
		}
	}

	for _, e := range allDirectionalEdges(g) {
		i := indices[e.From]
		j := indices[e.To]

		if e.Weight < data[i][j] {
			data[i][j] = e.Weight
		}
	}

	return Matrix{Indices: indices, Data: data}
}

// MatrixMultiply multiplies two matrices.
func MatrixMultiply(a, b Matrix) Matrix {
	n := len(a.Data)
	data := make([][]float64, n)

	for i := range n {
		data[i] = make([]float64, n)

		for j := range n {
			data[i][j] = math.Inf(1)

			for k := range n {
				if math.IsInf(a.Data[i][k], 1) || math.IsInf(b.Data[k][j], 1) {
					continue
				}

				val := a.Data[i][k] + b.Data[k][j]

				if val < data[i][j] {
					data[i][j] = val
				}
			}
		}
	}

	indices := make(map[string]int, len(a.Indices))
	maps.Copy(indices, a.Indices)

	return Matrix{Indices: indices, Data: data}
}

// MatrixPower raises a matrix to the given power using repeated squaring.
func MatrixPower(m Matrix, power int) Matrix {
	n := len(m.Data)

	result := Matrix{
		Indices: make(map[string]int, len(m.Indices)),
		Data:    make([][]float64, n),
	}

	maps.Copy(result.Indices, m.Indices)

	for i := range n {
		result.Data[i] = make([]float64, n)

		for j := range n {
			if i == j {
				result.Data[i][j] = 0
			} else {
				result.Data[i][j] = math.Inf(1)
			}
		}
	}

	base := m

	for power > 0 {
		if power%2 == 1 {
			result = MatrixMultiply(result, base)
		}

		base = MatrixMultiply(base, base)
		power /= 2
	}

	return result
}

// TransitiveClosure computes the transitive closure of a graph.
// Returns a map where result[u][v] is true if v is reachable from u.
func TransitiveClosure(g Graph) map[string]map[string]bool {
	nodes := g.Nodes()
	nodeIDs := make([]string, len(nodes))

	for i, n := range nodes {
		nodeIDs[i] = n.ID
	}

	slices.Sort(nodeIDs)

	reach := make(map[string]map[string]bool, len(nodeIDs))

	for _, u := range nodeIDs {
		reach[u] = make(map[string]bool, len(nodeIDs))
		reach[u][u] = true
	}

	for _, e := range allDirectionalEdges(g) {
		reach[e.From][e.To] = true
	}

	for _, k := range nodeIDs {
		for _, i := range nodeIDs {
			for _, j := range nodeIDs {
				if reach[i][k] && reach[k][j] {
					reach[i][j] = true
				}
			}
		}
	}

	return reach
}
