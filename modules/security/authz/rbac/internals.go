package rbac

// buildClosure expands the hierarchy edges into a transitive closure:
// closure[role] is the set of every ancestor role reachable from role
// via the inheritance edges. The returned map ALWAYS contains an
// entry for every role that appears as a child in the input, even if
// the closure is empty (defensive: callers do not have to nil-check).
//
// Cycles cause buildClosure to return (nil, ErrInheritanceCycle). The
// detection is depth-first: a node revisited while its DFS subtree is
// still on the stack signals a cycle.
func buildClosure(hierarchy map[string][]string) (map[string][]string, error) {
	closure := make(map[string][]string, len(hierarchy))

	for child := range hierarchy {
		visiting := map[string]bool{}
		visited := map[string]bool{}

		err := dfsClosure(child, hierarchy, visiting, visited)
		if err != nil {
			return nil, err
		}

		out := make([]string, 0, len(visited))
		for ancestor := range visited {
			if ancestor == child {
				continue
			}

			out = append(out, ancestor)
		}

		closure[child] = out
	}

	return closure, nil
}

// dfsClosure walks the inheritance DAG depth-first. visiting tracks
// the current recursion path (cycle detection); visited collects
// every node reached from the original root.
func dfsClosure(node string, hierarchy map[string][]string, visiting, visited map[string]bool) error {
	if visiting[node] {
		return ErrInheritanceCycle
	}

	if visited[node] {
		return nil
	}

	visiting[node] = true
	visited[node] = true

	for _, parent := range hierarchy[node] {
		err := dfsClosure(parent, hierarchy, visiting, visited)
		if err != nil {
			return err
		}
	}

	visiting[node] = false

	return nil
}
