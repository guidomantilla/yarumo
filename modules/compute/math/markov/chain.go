package markov

import (
	"math"
	rand "math/rand/v2"
	"slices"

	"github.com/guidomantilla/yarumo/compute/math/graph"
	"github.com/guidomantilla/yarumo/compute/math/stats"
)

// NewChain creates a new Markov chain from states and a transition probability matrix.
// The matrix[i][j] represents the probability of transitioning from states[i] to states[j].
// Each row must sum to 1 and all probabilities must be non-negative.
func NewChain(states []string, matrix [][]float64) (*Chain, error) {
	if len(states) == 0 {
		return nil, ErrMarkov(ErrEmptyChain)
	}

	n := len(states)
	index := make(map[string]int, n)

	for i, s := range states {
		_, exists := index[s]
		if exists {
			return nil, ErrMarkov(ErrDuplicateState)
		}

		index[s] = i
	}

	if len(matrix) != n {
		return nil, ErrMarkov(ErrInvalidMatrix)
	}

	for i := range n {
		if len(matrix[i]) != n {
			return nil, ErrMarkov(ErrInvalidMatrix)
		}

		sum := 0.0

		for j := range n {
			if matrix[i][j] < 0 {
				return nil, ErrMarkov(ErrInvalidProbability)
			}

			sum += matrix[i][j]
		}

		if math.Abs(sum-1.0) > epsilon {
			return nil, ErrMarkov(ErrInvalidRow)
		}
	}

	g := graph.NewDirected()

	for _, s := range states {
		_ = g.AddNode(graph.Node{ID: s})
	}

	for i := range n {
		for j := range n {
			if matrix[i][j] > 0 {
				_ = g.AddEdge(graph.Edge{
					ID:     states[i] + "\x00" + states[j],
					From:   states[i],
					To:     states[j],
					Weight: matrix[i][j],
				})
			}
		}
	}

	mat := make([][]float64, n)

	for i := range mat {
		mat[i] = make([]float64, n)
		copy(mat[i], matrix[i])
	}

	order := make([]string, n)
	copy(order, states)

	return &Chain{
		graph:  g,
		states: order,
		index:  index,
		matrix: mat,
	}, nil
}

// P returns the transition probability from one state to another.
func (c *Chain) P(from, to string) (float64, error) {
	fi, fromExists := c.index[from]
	if !fromExists {
		return 0, ErrMarkov(ErrStateNotFound)
	}

	ti, toExists := c.index[to]
	if !toExists {
		return 0, ErrMarkov(ErrStateNotFound)
	}

	return c.matrix[fi][ti], nil
}

// States returns all state names sorted alphabetically.
func (c *Chain) States() []string {
	result := make([]string, len(c.states))
	copy(result, c.states)
	slices.Sort(result)

	return result
}

// Graph returns a clone of the internal directed graph for analysis.
func (c *Chain) Graph() *graph.Directed {
	return c.graph.CloneDirected()
}

// StepN returns the probability distribution after n transitions from the initial state.
func (c *Chain) StepN(initial string, n int) (stats.Distribution, error) {
	idx, exists := c.index[initial]
	if !exists {
		return nil, ErrMarkov(ErrStateNotFound)
	}

	size := len(c.states)

	dist := make([]float64, size)
	dist[idx] = 1.0

	for range n {
		next := make([]float64, size)

		for i := range size {
			for j := range size {
				next[j] += dist[i] * c.matrix[i][j]
			}
		}

		dist = next
	}

	result := make(stats.Distribution)

	for i, p := range dist {
		if p > 0 {
			result[stats.Outcome(c.states[i])] = stats.Prob(p)
		}
	}

	return result, nil
}

// Simulate generates a random walk of n steps from the initial state.
// The returned slice contains n+1 elements: the initial state followed by n transitions.
func (c *Chain) Simulate(initial string, n int, rng *rand.Rand) ([]string, error) {
	_, exists := c.index[initial]
	if !exists {
		return nil, ErrMarkov(ErrStateNotFound)
	}

	path := make([]string, 0, n+1)
	current := initial
	path = append(path, current)

	size := len(c.states)

	for range n {
		idx := c.index[current]
		r := rng.Float64()
		cumulative := 0.0
		next := c.states[size-1]

		for j := range size {
			cumulative += c.matrix[idx][j]

			if r < cumulative {
				next = c.states[j]

				break
			}
		}

		current = next
		path = append(path, current)
	}

	return path, nil
}
