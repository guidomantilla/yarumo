package markov

import (
	"github.com/guidomantilla/yarumo/compute/math/graph"
)

// Classify returns the classification of each state in the chain.
func (c *Chain) Classify() map[string]StateClass {
	n := len(c.states)
	result := make(map[string]StateClass, n)

	// Find SCCs.
	sccs := graph.StronglyConnectedComponents(c.graph)

	// Build SCC membership.
	sccOf := make(map[string]int)

	for i, scc := range sccs {
		for _, s := range scc {
			sccOf[s] = i
		}
	}

	// Check if SCC has outgoing edges to other SCCs.
	sccHasOut := make([]bool, len(sccs))

	for i := range n {
		for j := range n {
			if c.matrix[i][j] > 0 && sccOf[c.states[i]] != sccOf[c.states[j]] {
				sccHasOut[sccOf[c.states[i]]] = true
			}
		}
	}

	for i := range n {
		s := c.states[i]

		if c.matrix[i][i] == 1.0 {
			result[s] = Absorbing

			continue
		}

		if sccHasOut[sccOf[s]] {
			result[s] = Transient
		} else {
			result[s] = Recurrent
		}
	}

	return result
}

// IsAbsorbing reports whether the given state is absorbing (P(i,i) = 1).
func (c *Chain) IsAbsorbing(state string) (bool, error) {
	idx, exists := c.index[state]
	if !exists {
		return false, ErrMarkov(ErrStateNotFound)
	}

	return c.matrix[idx][idx] == 1.0, nil
}

// IsIrreducible reports whether the chain has a single communicating class.
func (c *Chain) IsIrreducible() bool {
	sccs := graph.StronglyConnectedComponents(c.graph)

	return len(sccs) == 1
}

// Period returns the period of a state.
// The period is the GCD of all cycle lengths passing through the state.
// Returns 0 if no cycles pass through the state.
func (c *Chain) Period(state string) (int, error) {
	idx, exists := c.index[state]
	if !exists {
		return 0, ErrMarkov(ErrStateNotFound)
	}

	n := len(c.states)

	dist := make([]int, n)

	for i := range dist {
		dist[i] = -1
	}

	dist[idx] = 0

	queue := []int{idx}
	period := 0
	hasCycle := false

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for j := range n {
			if c.matrix[curr][j] == 0 {
				continue
			}

			if j == idx {
				hasCycle = true
			}

			if dist[j] == -1 {
				dist[j] = dist[curr] + 1
				queue = append(queue, j)
			} else {
				period = gcd(period, dist[curr]+1-dist[j])
			}
		}
	}

	if !hasCycle {
		return 0, nil
	}

	return period, nil
}

// IsErgodic reports whether the chain is irreducible and aperiodic.
func (c *Chain) IsErgodic() bool {
	if !c.IsIrreducible() {
		return false
	}

	// For an irreducible chain, all states have the same period.
	p, _ := c.Period(c.states[0])

	return p == 1
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}

	return a
}
