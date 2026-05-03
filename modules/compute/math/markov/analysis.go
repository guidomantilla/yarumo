package markov

import (
	"github.com/guidomantilla/yarumo/compute/math/stats"
)

// Steady returns the stationary distribution of an irreducible chain.
func (c *Chain) Steady() (stats.Distribution, error) {
	if !c.IsIrreducible() {
		return nil, ErrMarkov(ErrNotIrreducible)
	}

	n := len(c.states)

	// Solve (P^T - I) π = 0 with constraint Σ π_i = 1.
	// Replace the last equation with the normalization constraint.
	a := make([][]float64, n)
	b := make([]float64, n)

	for i := range n {
		a[i] = make([]float64, n)

		if i < n-1 {
			for j := range n {
				a[i][j] = c.matrix[j][i] // P^T.

				if i == j {
					a[i][j] -= 1.0 // - I.
				}
			}
		} else {
			for j := range n {
				a[i][j] = 1.0
			}

			b[i] = 1.0
		}
	}

	x, err := solveLinearSystem(a, b)
	if err != nil {
		return nil, err
	}

	result := make(stats.Distribution)

	for i, p := range x {
		if p > 0 {
			result[stats.Outcome(c.states[i])] = stats.Prob(p)
		}
	}

	return result, nil
}

// MeanFirstPassage returns the expected number of steps to reach the target state
// from the source state for the first time.
func (c *Chain) MeanFirstPassage(from, to string) (float64, error) {
	fromIdx, fromExists := c.index[from]
	if !fromExists {
		return 0, ErrMarkov(ErrStateNotFound)
	}

	toIdx, toExists := c.index[to]
	if !toExists {
		return 0, ErrMarkov(ErrStateNotFound)
	}

	if fromIdx == toIdx {
		return 0, nil
	}

	n := len(c.states)
	size := n - 1

	// Map reduced indices to original indices (excluding target).
	origIdx := make([]int, 0, size)
	fromReduced := 0

	for i := range n {
		if i != toIdx {
			if i == fromIdx {
				fromReduced = len(origIdx)
			}

			origIdx = append(origIdx, i)
		}
	}

	// Build system: (I - Q) h = 1 where Q excludes target state.
	a := make([][]float64, size)
	b := make([]float64, size)

	for i := range size {
		a[i] = make([]float64, size)
		oi := origIdx[i]

		for j := range size {
			oj := origIdx[j]

			if i == j {
				a[i][j] = 1.0 - c.matrix[oi][oj]
			} else {
				a[i][j] = -c.matrix[oi][oj]
			}
		}

		b[i] = 1.0
	}

	x, err := solveLinearSystem(a, b)
	if err != nil {
		return 0, err
	}

	return x[fromReduced], nil
}

// AbsorptionProbabilities returns the probability of being absorbed by each absorbing
// state from each transient state.
// The result maps transient state → absorbing state → probability.
func (c *Chain) AbsorptionProbabilities() (map[string]map[string]float64, error) {
	classes := c.Classify()

	var transient []int

	var absorbing []int

	for i, s := range c.states {
		switch classes[s] {
		case Transient:
			transient = append(transient, i)
		case Absorbing:
			absorbing = append(absorbing, i)
		case Recurrent:
		}
	}

	if len(absorbing) == 0 {
		return nil, ErrMarkov(ErrNoAbsorbingStates)
	}

	if len(transient) == 0 {
		return make(map[string]map[string]float64), nil
	}

	nt := len(transient)

	// Build (I - Q) where Q is the transient-to-transient submatrix.
	iq := make([][]float64, nt)

	for i := range iq {
		iq[i] = make([]float64, nt)

		for j := range iq[i] {
			if i == j {
				iq[i][j] = 1.0 - c.matrix[transient[i]][transient[j]]
			} else {
				iq[i][j] = -c.matrix[transient[i]][transient[j]]
			}
		}
	}

	result := make(map[string]map[string]float64, nt)

	for _, ti := range transient {
		result[c.states[ti]] = make(map[string]float64, len(absorbing))
	}

	// Solve (I - Q) B_j = R_j for each absorbing state j.
	for _, aj := range absorbing {
		b := make([]float64, nt)

		for i, ti := range transient {
			b[i] = c.matrix[ti][aj]
		}

		x, err := solveLinearSystem(iq, b)
		if err != nil {
			return nil, err
		}

		for i, ti := range transient {
			result[c.states[ti]][c.states[aj]] = x[i]
		}
	}

	return result, nil
}

// MeanAbsorptionTime returns the expected number of steps until the chain
// is absorbed into an absorbing state, starting from the given transient state.
func (c *Chain) MeanAbsorptionTime(from string) (float64, error) {
	fromIdx, exists := c.index[from]
	if !exists {
		return 0, ErrMarkov(ErrStateNotFound)
	}

	classes := c.Classify()

	if classes[from] != Transient {
		return 0, ErrMarkov(ErrNotTransient)
	}

	var absorbing []int

	var transient []int

	fromReduced := 0

	for i, s := range c.states {
		switch classes[s] {
		case Absorbing:
			absorbing = append(absorbing, i)
		case Transient:
			if i == fromIdx {
				fromReduced = len(transient)
			}

			transient = append(transient, i)
		case Recurrent:
		}
	}

	if len(absorbing) == 0 {
		return 0, ErrMarkov(ErrNoAbsorbingStates)
	}

	nt := len(transient)

	// Build (I - Q) t = 1.
	a := make([][]float64, nt)
	b := make([]float64, nt)

	for i := range a {
		a[i] = make([]float64, nt)

		for j := range a[i] {
			if i == j {
				a[i][j] = 1.0 - c.matrix[transient[i]][transient[j]]
			} else {
				a[i][j] = -c.matrix[transient[i]][transient[j]]
			}
		}

		b[i] = 1.0
	}

	x, err := solveLinearSystem(a, b)
	if err != nil {
		return 0, err
	}

	return x[fromReduced], nil
}
