// Package explain provides explanation trace types for MCDM analysis.
package explain

// RankEntry records a single alternative's ranking details.
type RankEntry struct {
	Alternative int
	Score       float64
	Rank        int
}

// Trace records the MCDM computation for explainability.
type Trace struct {
	Method   string      // "AHP", "TOPSIS", "AHP+TOPSIS".
	Criteria []string    // Criterion names.
	Weights  []float64   // Criterion weights.
	Rankings []RankEntry // Sorted by rank.
}

// NewTrace creates a new MCDM trace.
func NewTrace(method string, criteria []string, weights []float64) Trace {
	return Trace{
		Method:   method,
		Criteria: criteria,
		Weights:  weights,
	}
}

// AddRanking appends a rank entry.
func (t Trace) AddRanking(entry RankEntry) Trace {
	t.Rankings = append(t.Rankings, entry)
	return t
}
