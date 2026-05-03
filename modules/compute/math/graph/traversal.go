package graph

// BFS performs a breadth-first search from the start node, calling visit for each visited node.
// Returns an error if the start node is not found.
func BFS(g Graph, start string, visit func(id string)) error {
	if !g.HasNode(start) {
		return ErrGraph(ErrNodeNotFound)
	}

	visited := make(map[string]bool)
	queue := []string{start}
	visited[start] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		visit(curr)

		neighbors, _ := g.Neighbors(curr)

		for _, nb := range neighbors {
			if !visited[nb] {
				visited[nb] = true
				queue = append(queue, nb)
			}
		}
	}

	return nil
}

// DFS performs a depth-first search from the start node, calling visit for each visited node.
// Returns an error if the start node is not found.
func DFS(g Graph, start string, visit func(id string)) error {
	if !g.HasNode(start) {
		return ErrGraph(ErrNodeNotFound)
	}

	visited := make(map[string]bool)

	var dfs func(string)
	dfs = func(id string) {
		visited[id] = true
		visit(id)

		neighbors, _ := g.Neighbors(id)

		for _, nb := range neighbors {
			if !visited[nb] {
				dfs(nb)
			}
		}
	}

	dfs(start)

	return nil
}

// BFSAll performs BFS from the start node and returns all visited node IDs in BFS order.
func BFSAll(g Graph, start string) ([]string, error) {
	var result []string

	err := BFS(g, start, func(id string) {
		result = append(result, id)
	})

	return result, err
}

// DFSAll performs DFS from the start node and returns all visited node IDs in DFS order.
func DFSAll(g Graph, start string) ([]string, error) {
	var result []string

	err := DFS(g, start, func(id string) {
		result = append(result, id)
	})

	return result, err
}
