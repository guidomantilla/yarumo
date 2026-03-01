package explain

import (
	"testing"

	"github.com/guidomantilla/yarumo/maths/probability"
)

func TestNewTrace(t *testing.T) {
	t.Parallel()

	tr := NewTrace("X", probability.Assignment{"Y": "true"})

	if tr.Query != "X" {
		t.Fatalf("expected query X, got %s", string(tr.Query))
	}

	if tr.Evidence["Y"] != "true" {
		t.Fatalf("expected evidence Y=true")
	}

	if len(tr.Steps) != 0 {
		t.Fatalf("expected empty steps, got %d", len(tr.Steps))
	}
}

func TestTrace_AddStep(t *testing.T) {
	t.Parallel()

	tr := NewTrace("X", nil)
	tr = tr.AddStep(Step{Number: 1, Phase: Initialize, Message: "init"})
	tr = tr.AddStep(Step{Number: 2, Phase: Complete, Message: "done"})

	if len(tr.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(tr.Steps))
	}

	if tr.Steps[0].Number != 1 {
		t.Fatalf("expected step 1, got %d", tr.Steps[0].Number)
	}

	if tr.Steps[1].Number != 2 {
		t.Fatalf("expected step 2, got %d", tr.Steps[1].Number)
	}
}

func TestTrace_AddPosterior(t *testing.T) {
	t.Parallel()

	tr := NewTrace("X", nil)
	tr = tr.AddPosterior(Posterior{
		Variable:     "X",
		Distribution: probability.Distribution{"true": 0.7, "false": 0.3},
	})

	if len(tr.Posteriors) != 1 {
		t.Fatalf("expected 1 posterior, got %d", len(tr.Posteriors))
	}

	if tr.Posteriors[0].Variable != "X" {
		t.Fatalf("expected X, got %s", string(tr.Posteriors[0].Variable))
	}
}
