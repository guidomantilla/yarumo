package graph

// GraphColoring returns a greedy coloring of the graph.
// Returns a map from node ID to color number (0-indexed).
func GraphColoring(g Graph) map[string]int {
	nodes := g.Nodes()
	color := make(map[string]int, len(nodes))

	for _, n := range nodes {
		color[n.ID] = -1
	}

	for _, n := range nodes {
		used := make(map[int]bool)
		neighbors, _ := g.Neighbors(n.ID)

		for _, nb := range neighbors {
			if color[nb] >= 0 {
				used[color[nb]] = true
			}
		}

		c := 0

		for used[c] {
			c++
		}

		color[n.ID] = c
	}

	return color
}

// ChromaticNumber returns the number of colors used by the greedy coloring.
func ChromaticNumber(g Graph) int {
	coloring := GraphColoring(g)
	maxColor := -1

	for _, c := range coloring {
		if c > maxColor {
			maxColor = c
		}
	}

	return maxColor + 1
}
