package graph

import (
	"math"
)

// DegreeCentrality returns the degree centrality of each node.
func DegreeCentrality(g Graph) map[string]float64 {
	nodes := g.Nodes()
	n := len(nodes)
	result := make(map[string]float64, n)

	if n <= 1 {
		for _, node := range nodes {
			result[node.ID] = 0
		}

		return result
	}

	for _, node := range nodes {
		deg, _ := g.Degree(node.ID)
		result[node.ID] = float64(deg) / float64(n-1)
	}

	return result
}

// BetweennessCentrality returns the betweenness centrality of each node.
func BetweennessCentrality(g Graph) map[string]float64 {
	nodes := g.Nodes()
	centrality := make(map[string]float64, len(nodes))

	for _, n := range nodes {
		centrality[n.ID] = 0
	}

	for _, s := range nodes {
		stack := make([]string, 0)
		pred := make(map[string][]string)
		sigma := make(map[string]float64)
		dist := make(map[string]int)

		for _, n := range nodes {
			pred[n.ID] = nil
			sigma[n.ID] = 0
			dist[n.ID] = -1
		}

		sigma[s.ID] = 1
		dist[s.ID] = 0
		queue := []string{s.ID}

		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			stack = append(stack, v)

			neighbors, _ := g.Neighbors(v)

			for _, w := range neighbors {
				if dist[w] < 0 {
					queue = append(queue, w)
					dist[w] = dist[v] + 1
				}

				if dist[w] == dist[v]+1 {
					sigma[w] += sigma[v]
					pred[w] = append(pred[w], v)
				}
			}
		}

		delta := make(map[string]float64)

		for i := len(stack) - 1; i >= 0; i-- {
			w := stack[i]

			for _, v := range pred[w] {
				delta[v] += (sigma[v] / sigma[w]) * (1 + delta[w])
			}

			if w != s.ID {
				centrality[w] += delta[w]
			}
		}
	}

	// For undirected graphs, each pair is counted twice; normalize by dividing by 2.
	if !g.IsDirected() {
		for id := range centrality {
			centrality[id] /= 2
		}
	}

	return centrality
}

// ClosenessCentrality returns the closeness centrality of each node.
func ClosenessCentrality(g Graph) map[string]float64 {
	nodes := g.Nodes()
	n := len(nodes)
	result := make(map[string]float64, n)

	if n <= 1 {
		for _, node := range nodes {
			result[node.ID] = 0
		}

		return result
	}

	dist := AllPairsShortestPath(g)

	for _, node := range nodes {
		totalDist := 0.0
		reachable := 0

		for _, other := range nodes {
			if node.ID == other.ID {
				continue
			}

			d := dist[node.ID][other.ID]

			if !math.IsInf(d, 1) {
				totalDist += d
				reachable++
			}
		}

		if reachable > 0 && totalDist > 0 {
			result[node.ID] = float64(reachable) / totalDist
		}
	}

	return result
}

// PageRank computes PageRank centrality with the given damping factor and iterations.
func PageRank(g *Directed, damping float64, iterations int) map[string]float64 {
	nodes := g.Nodes()
	n := len(nodes)
	rank := make(map[string]float64, n)

	if n == 0 {
		return rank
	}

	initial := 1.0 / float64(n)

	for _, node := range nodes {
		rank[node.ID] = initial
	}

	for range iterations {
		newRank := make(map[string]float64, n)

		for _, node := range nodes {
			newRank[node.ID] = (1 - damping) / float64(n)
		}

		for _, node := range nodes {
			outDeg := len(g.out[node.ID])
			if outDeg == 0 {
				share := rank[node.ID] / float64(n)

				for _, other := range nodes {
					newRank[other.ID] += damping * share
				}
			} else {
				share := rank[node.ID] / float64(outDeg)

				for _, eid := range g.out[node.ID] {
					edge := g.edges[eid]
					newRank[edge.To] += damping * share
				}
			}
		}

		rank = newRank
	}

	return rank
}
