package explain

import "testing"

func TestRankEntry_String(t *testing.T) {
	t.Parallel()

	t.Run("basic entry", func(t *testing.T) {
		t.Parallel()

		entry := RankEntry{Alternative: 0, Score: 0.85, Rank: 1}
		got := entry.String()

		expected := "#1 alternative 0 (score: 0.850)"
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("zero score", func(t *testing.T) {
		t.Parallel()

		entry := RankEntry{Alternative: 2, Score: 0, Rank: 3}
		got := entry.String()

		expected := "#3 alternative 2 (score: 0.000)"
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("high rank number", func(t *testing.T) {
		t.Parallel()

		entry := RankEntry{Alternative: 10, Score: 0.123, Rank: 5}
		got := entry.String()

		expected := "#5 alternative 10 (score: 0.123)"
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})
}

func TestTrace_String(t *testing.T) {
	t.Parallel()

	t.Run("full trace", func(t *testing.T) {
		t.Parallel()

		tr := NewTrace("AHP", []string{"cost", "quality", "speed"}, []float64{0.5, 0.3, 0.2})
		tr = tr.AddRanking(RankEntry{Alternative: 1, Score: 0.85, Rank: 1})
		tr = tr.AddRanking(RankEntry{Alternative: 0, Score: 0.65, Rank: 2})

		got := tr.String()

		expected := "method: AHP\ncriteria: cost (0.500), quality (0.300), speed (0.200)\nrankings:\n  #1 alternative 1 (score: 0.850)\n  #2 alternative 0 (score: 0.650)"
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("method only", func(t *testing.T) {
		t.Parallel()

		tr := NewTrace("TOPSIS", nil, nil)

		got := tr.String()

		expected := "method: TOPSIS"
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("criteria without weights", func(t *testing.T) {
		t.Parallel()

		tr := NewTrace("AHP+TOPSIS", []string{"cost", "quality"}, nil)

		got := tr.String()

		expected := "method: AHP+TOPSIS\ncriteria: cost, quality"
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("criteria with partial weights", func(t *testing.T) {
		t.Parallel()

		tr := NewTrace("AHP", []string{"cost", "quality", "speed"}, []float64{0.5})

		got := tr.String()

		expected := "method: AHP\ncriteria: cost (0.500), quality, speed"
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("empty rankings", func(t *testing.T) {
		t.Parallel()

		tr := NewTrace("TOPSIS", []string{"price"}, []float64{1.0})

		got := tr.String()

		expected := "method: TOPSIS\ncriteria: price (1.000)"
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})
}

func TestNewTrace(t *testing.T) {
	t.Parallel()

	tr := NewTrace("AHP", []string{"a", "b"}, []float64{0.6, 0.4})

	if tr.Method != "AHP" {
		t.Fatalf("expected method AHP, got %s", tr.Method)
	}

	if len(tr.Criteria) != 2 {
		t.Fatalf("expected 2 criteria, got %d", len(tr.Criteria))
	}

	if len(tr.Weights) != 2 {
		t.Fatalf("expected 2 weights, got %d", len(tr.Weights))
	}

	if tr.Rankings != nil {
		t.Fatalf("expected nil rankings, got %v", tr.Rankings)
	}
}

func TestTrace_AddRanking(t *testing.T) {
	t.Parallel()

	tr := NewTrace("TOPSIS", []string{"x"}, []float64{1.0})

	tr = tr.AddRanking(RankEntry{Alternative: 0, Score: 0.9, Rank: 1})
	tr = tr.AddRanking(RankEntry{Alternative: 1, Score: 0.7, Rank: 2})

	if len(tr.Rankings) != 2 {
		t.Fatalf("expected 2 rankings, got %d", len(tr.Rankings))
	}

	if tr.Rankings[0].Alternative != 0 {
		t.Fatalf("expected first ranking alternative 0, got %d", tr.Rankings[0].Alternative)
	}

	if tr.Rankings[1].Rank != 2 {
		t.Fatalf("expected second ranking rank 2, got %d", tr.Rankings[1].Rank)
	}
}

func Test_formatFloat(t *testing.T) {
	t.Parallel()

	t.Run("zero", func(t *testing.T) {
		t.Parallel()

		got := formatFloat(0)
		if got != "0.000" {
			t.Fatalf("expected 0.000, got %s", got)
		}
	})

	t.Run("simple decimal", func(t *testing.T) {
		t.Parallel()

		got := formatFloat(0.5)
		if got != "0.500" {
			t.Fatalf("expected 0.500, got %s", got)
		}
	})

	t.Run("three decimals", func(t *testing.T) {
		t.Parallel()

		got := formatFloat(0.123)
		if got != "0.123" {
			t.Fatalf("expected 0.123, got %s", got)
		}
	})

	t.Run("integer value", func(t *testing.T) {
		t.Parallel()

		got := formatFloat(1.0)
		if got != "1.000" {
			t.Fatalf("expected 1.000, got %s", got)
		}
	})

	t.Run("negative value", func(t *testing.T) {
		t.Parallel()

		got := formatFloat(-1.5)
		if got != "-1.500" {
			t.Fatalf("expected -1.500, got %s", got)
		}
	})
}

func Test_formatCriteria(t *testing.T) {
	t.Parallel()

	t.Run("names with weights", func(t *testing.T) {
		t.Parallel()

		got := formatCriteria([]string{"cost", "quality"}, []float64{0.6, 0.4})

		expected := "cost (0.600), quality (0.400)"
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("names without weights", func(t *testing.T) {
		t.Parallel()

		got := formatCriteria([]string{"cost", "quality"}, nil)

		expected := "cost, quality"
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("single criterion", func(t *testing.T) {
		t.Parallel()

		got := formatCriteria([]string{"price"}, []float64{1.0})

		expected := "price (1.000)"
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		got := formatCriteria(nil, nil)
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})
}
