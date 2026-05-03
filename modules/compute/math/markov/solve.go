package markov

import "math"

const epsilon = 1e-9

// solveLinearSystem solves Ax = b using Gaussian elimination with partial pivoting.
func solveLinearSystem(a [][]float64, b []float64) ([]float64, error) {
	n := len(b)

	// Create augmented matrix [A|b].
	aug := make([][]float64, n)

	for i := range aug {
		aug[i] = make([]float64, n+1)
		copy(aug[i], a[i])
		aug[i][n] = b[i]
	}

	// Forward elimination with partial pivoting.
	for col := range n {
		maxRow := col
		maxVal := math.Abs(aug[col][col])

		for row := col + 1; row < n; row++ {
			v := math.Abs(aug[row][col])

			if v > maxVal {
				maxVal = v
				maxRow = row
			}
		}

		if maxVal < epsilon {
			return nil, ErrMarkov(ErrSingularMatrix)
		}

		aug[col], aug[maxRow] = aug[maxRow], aug[col]

		for row := col + 1; row < n; row++ {
			factor := aug[row][col] / aug[col][col]

			for j := col; j <= n; j++ {
				aug[row][j] -= factor * aug[col][j]
			}
		}
	}

	// Back substitution.
	x := make([]float64, n)

	for i := n - 1; i >= 0; i-- {
		x[i] = aug[i][n]

		for j := i + 1; j < n; j++ {
			x[i] -= aug[i][j] * x[j]
		}

		x[i] /= aug[i][i]
	}

	return x, nil
}
