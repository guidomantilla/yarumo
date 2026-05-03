package temporal

import (
	"errors"
	"testing"
	"time"
)

func baseTime() time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
}

// --- ResponseWithin ---

func TestResponseWithin_allResponded(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "request", Time: base},
		{Label: "response", Time: base.Add(1 * time.Hour)},
	}

	ok, violations := ResponseWithin(trace, "request", "response", 2*time.Hour)

	if !ok {
		t.Fatal("expected true")
	}

	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %d", len(violations))
	}
}

func TestResponseWithin_violated(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "request", Time: base},
		{Label: "response", Time: base.Add(3 * time.Hour)},
	}

	ok, violations := ResponseWithin(trace, "request", "response", 2*time.Hour)

	if ok {
		t.Fatal("expected false")
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
}

func TestResponseWithin_multipleRequests(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "request", Time: base},
		{Label: "response", Time: base.Add(1 * time.Hour)},
		{Label: "request", Time: base.Add(2 * time.Hour)},
		{Label: "response", Time: base.Add(5 * time.Hour)},
	}

	ok, violations := ResponseWithin(trace, "request", "response", 2*time.Hour)

	if ok {
		t.Fatal("expected false")
	}

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	if violations[0].TriggerEvent.Time != base.Add(2*time.Hour) {
		t.Fatalf("expected violation for second request, got %v", violations[0].TriggerEvent.Time)
	}
}

func TestResponseWithin_noTriggers(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "other", Time: base},
		{Label: "response", Time: base.Add(1 * time.Hour)},
	}

	ok, violations := ResponseWithin(trace, "request", "response", 2*time.Hour)

	if !ok {
		t.Fatal("expected true")
	}

	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %d", len(violations))
	}
}

// --- FrequencyWithin ---

func TestFrequencyWithin_detected(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "txn", Time: base},
		{Label: "txn", Time: base.Add(1 * time.Minute)},
		{Label: "txn", Time: base.Add(2 * time.Minute)},
	}

	got := FrequencyWithin(trace, "txn", 3, 5*time.Minute)

	if !got {
		t.Fatal("expected true")
	}
}

func TestFrequencyWithin_notDetected(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "txn", Time: base},
		{Label: "txn", Time: base.Add(30 * time.Minute)},
		{Label: "txn", Time: base.Add(60 * time.Minute)},
	}

	got := FrequencyWithin(trace, "txn", 3, 5*time.Minute)

	if got {
		t.Fatal("expected false")
	}
}

func TestFrequencyWithin_exactThreshold(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "txn", Time: base},
		{Label: "txn", Time: base.Add(2 * time.Minute)},
		{Label: "txn", Time: base.Add(5 * time.Minute)},
	}

	got := FrequencyWithin(trace, "txn", 3, 5*time.Minute)

	if !got {
		t.Fatal("expected true")
	}
}

func TestFrequencyWithin_insufficientEvents(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "txn", Time: base},
		{Label: "txn", Time: base.Add(1 * time.Minute)},
	}

	got := FrequencyWithin(trace, "txn", 3, 5*time.Minute)

	if got {
		t.Fatal("expected false")
	}
}

// --- Eventually ---

func TestEventually_found(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "start", Time: base},
		{Label: "complete", Time: base.Add(1 * time.Hour)},
	}

	got := Eventually(trace, "complete")

	if !got {
		t.Fatal("expected true")
	}
}

func TestEventually_notFound(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "start", Time: base},
		{Label: "progress", Time: base.Add(1 * time.Hour)},
	}

	got := Eventually(trace, "complete")

	if got {
		t.Fatal("expected false")
	}
}

func TestEventually_emptyTrace(t *testing.T) {
	t.Parallel()

	got := Eventually(Trace{}, "complete")

	if got {
		t.Fatal("expected false")
	}
}

// --- Before ---

func TestBefore_satisfied(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "a", Time: base.Add(1 * time.Hour)},
		{Label: "b", Time: base.Add(2 * time.Hour)},
	}

	got := Before(trace, "a", "b")

	if !got {
		t.Fatal("expected true")
	}
}

func TestBefore_violated(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "b", Time: base.Add(1 * time.Hour)},
		{Label: "a", Time: base.Add(2 * time.Hour)},
	}

	got := Before(trace, "a", "b")

	if got {
		t.Fatal("expected false")
	}
}

func TestBefore_noBPresent(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "a", Time: base.Add(1 * time.Hour)},
		{Label: "c", Time: base.Add(2 * time.Hour)},
	}

	got := Before(trace, "a", "b")

	if !got {
		t.Fatal("expected true when b is absent")
	}
}

// --- Elapsed ---

func TestElapsed_basic(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "start", Time: base},
		{Label: "end", Time: base.Add(3 * time.Hour)},
	}

	dur, err := Elapsed(trace, "start", "end")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dur != 3*time.Hour {
		t.Fatalf("expected 3h, got %s", dur)
	}
}

func TestElapsed_fromNotFound(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "end", Time: base.Add(3 * time.Hour)},
	}

	_, err := Elapsed(trace, "start", "end")

	if !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestElapsed_toNotFound(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "start", Time: base},
	}

	_, err := Elapsed(trace, "start", "end")

	if !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound, got %v", err)
	}
}

func TestElapsed_negative(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "to", Time: base},
		{Label: "from", Time: base.Add(3 * time.Hour)},
	}

	dur, err := Elapsed(trace, "from", "to")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dur >= 0 {
		t.Fatalf("expected negative duration, got %s", dur)
	}
}

// --- Always ---

func TestAlways(t *testing.T) {
	t.Parallel()

	base := baseTime()

	t.Run("all_match", func(t *testing.T) {
		t.Parallel()

		trace := Trace{
			{Label: "ok", Time: base},
			{Label: "ok", Time: base.Add(1 * time.Hour)},
			{Label: "ok", Time: base.Add(2 * time.Hour)},
		}

		got := Always(trace, func(ev Event) bool { return ev.Label == "ok" })

		if !got {
			t.Fatal("expected true")
		}
	})

	t.Run("one_fails", func(t *testing.T) {
		t.Parallel()

		trace := Trace{
			{Label: "ok", Time: base},
			{Label: "bad", Time: base.Add(1 * time.Hour)},
			{Label: "ok", Time: base.Add(2 * time.Hour)},
		}

		got := Always(trace, func(ev Event) bool { return ev.Label == "ok" })

		if got {
			t.Fatal("expected false")
		}
	})

	t.Run("empty_trace", func(t *testing.T) {
		t.Parallel()

		got := Always(Trace{}, func(ev Event) bool { return false })

		if !got {
			t.Fatal("expected true for empty trace")
		}
	})
}

// --- Next ---

func TestNext(t *testing.T) {
	t.Parallel()

	base := baseTime()

	t.Run("next_matches", func(t *testing.T) {
		t.Parallel()

		trace := Trace{
			{Label: "trigger", Time: base},
			{Label: "expected", Time: base.Add(1 * time.Hour)},
		}

		got := Next(trace, "trigger", func(ev Event) bool { return ev.Label == "expected" })

		if !got {
			t.Fatal("expected true")
		}
	})

	t.Run("next_fails", func(t *testing.T) {
		t.Parallel()

		trace := Trace{
			{Label: "trigger", Time: base},
			{Label: "other", Time: base.Add(1 * time.Hour)},
		}

		got := Next(trace, "trigger", func(ev Event) bool { return ev.Label == "expected" })

		if got {
			t.Fatal("expected false")
		}
	})

	t.Run("event_last", func(t *testing.T) {
		t.Parallel()

		trace := Trace{
			{Label: "other", Time: base},
			{Label: "trigger", Time: base.Add(1 * time.Hour)},
		}

		got := Next(trace, "trigger", func(ev Event) bool { return true })

		if got {
			t.Fatal("expected false when event is last")
		}
	})

	t.Run("event_not_found", func(t *testing.T) {
		t.Parallel()

		trace := Trace{
			{Label: "a", Time: base},
			{Label: "b", Time: base.Add(1 * time.Hour)},
		}

		got := Next(trace, "trigger", func(ev Event) bool { return true })

		if got {
			t.Fatal("expected false when event not found")
		}
	})

	t.Run("empty_trace", func(t *testing.T) {
		t.Parallel()

		got := Next(Trace{}, "trigger", func(ev Event) bool { return true })

		if got {
			t.Fatal("expected false for empty trace")
		}
	})
}

// --- Until ---

func TestUntil(t *testing.T) {
	t.Parallel()

	base := baseTime()

	t.Run("b_after_as", func(t *testing.T) {
		t.Parallel()

		trace := Trace{
			{Label: "a", Time: base},
			{Label: "a", Time: base.Add(1 * time.Hour)},
			{Label: "b", Time: base.Add(2 * time.Hour)},
		}

		got := Until(trace, "a", "b")

		if !got {
			t.Fatal("expected true")
		}
	})

	t.Run("b_first", func(t *testing.T) {
		t.Parallel()

		trace := Trace{
			{Label: "b", Time: base},
			{Label: "a", Time: base.Add(1 * time.Hour)},
		}

		got := Until(trace, "a", "b")

		if !got {
			t.Fatal("expected true when b is first")
		}
	})

	t.Run("no_b", func(t *testing.T) {
		t.Parallel()

		trace := Trace{
			{Label: "a", Time: base},
			{Label: "a", Time: base.Add(1 * time.Hour)},
		}

		got := Until(trace, "a", "b")

		if got {
			t.Fatal("expected false when b never occurs")
		}
	})

	t.Run("interrupted", func(t *testing.T) {
		t.Parallel()

		trace := Trace{
			{Label: "a", Time: base},
			{Label: "c", Time: base.Add(1 * time.Hour)},
			{Label: "b", Time: base.Add(2 * time.Hour)},
		}

		got := Until(trace, "a", "b")

		if got {
			t.Fatal("expected false when a is interrupted")
		}
	})

	t.Run("empty_trace", func(t *testing.T) {
		t.Parallel()

		got := Until(Trace{}, "a", "b")

		if got {
			t.Fatal("expected false for empty trace")
		}
	})
}

// --- Release ---

func TestRelease(t *testing.T) {
	t.Parallel()

	base := baseTime()

	t.Run("a_and_b_together", func(t *testing.T) {
		t.Parallel()

		// When a == b, the first event satisfies both conditions.
		trace := Trace{
			{Label: "ab", Time: base},
			{Label: "c", Time: base.Add(1 * time.Hour)},
		}

		got := Release(trace, "ab", "ab")

		if !got {
			t.Fatal("expected true")
		}
	})

	t.Run("b_always", func(t *testing.T) {
		t.Parallel()

		trace := Trace{
			{Label: "b", Time: base},
			{Label: "b", Time: base.Add(1 * time.Hour)},
			{Label: "b", Time: base.Add(2 * time.Hour)},
		}

		got := Release(trace, "a", "b")

		if !got {
			t.Fatal("expected true when b holds forever")
		}
	})

	t.Run("b_breaks", func(t *testing.T) {
		t.Parallel()

		trace := Trace{
			{Label: "b", Time: base},
			{Label: "c", Time: base.Add(1 * time.Hour)},
		}

		got := Release(trace, "a", "b")

		if got {
			t.Fatal("expected false when b breaks without a")
		}
	})

	t.Run("empty_trace", func(t *testing.T) {
		t.Parallel()

		got := Release(Trace{}, "a", "b")

		if !got {
			t.Fatal("expected true for empty trace")
		}
	})
}

// --- Since ---

func TestSince(t *testing.T) {
	t.Parallel()

	base := baseTime()

	t.Run("b_then_as", func(t *testing.T) {
		t.Parallel()

		trace := Trace{
			{Label: "b", Time: base},
			{Label: "a", Time: base.Add(1 * time.Hour)},
			{Label: "a", Time: base.Add(2 * time.Hour)},
		}

		got := Since(trace, "a", "b")

		if !got {
			t.Fatal("expected true")
		}
	})

	t.Run("a_without_b", func(t *testing.T) {
		t.Parallel()

		trace := Trace{
			{Label: "a", Time: base},
			{Label: "a", Time: base.Add(1 * time.Hour)},
		}

		got := Since(trace, "a", "b")

		if got {
			t.Fatal("expected false when b never occurred")
		}
	})

	t.Run("interrupted", func(t *testing.T) {
		t.Parallel()

		trace := Trace{
			{Label: "b", Time: base},
			{Label: "c", Time: base.Add(1 * time.Hour)},
			{Label: "a", Time: base.Add(2 * time.Hour)},
		}

		got := Since(trace, "a", "b")

		if got {
			t.Fatal("expected false when a is interrupted going backward")
		}
	})

	t.Run("empty_trace", func(t *testing.T) {
		t.Parallel()

		got := Since(Trace{}, "a", "b")

		if got {
			t.Fatal("expected false for empty trace")
		}
	})
}

// --- Sequence ---

func TestSequence_inOrder(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "a", Time: base},
		{Label: "b", Time: base.Add(1 * time.Hour)},
		{Label: "c", Time: base.Add(2 * time.Hour)},
	}

	got := Sequence(trace, "a", "b", "c")

	if !got {
		t.Fatal("expected true")
	}
}

func TestSequence_outOfOrder(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "c", Time: base},
		{Label: "a", Time: base.Add(1 * time.Hour)},
		{Label: "b", Time: base.Add(2 * time.Hour)},
	}

	got := Sequence(trace, "a", "b", "c")

	if got {
		t.Fatal("expected false")
	}
}

func TestSequence_withGaps(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "a", Time: base},
		{Label: "x", Time: base.Add(1 * time.Hour)},
		{Label: "b", Time: base.Add(2 * time.Hour)},
		{Label: "y", Time: base.Add(3 * time.Hour)},
		{Label: "c", Time: base.Add(4 * time.Hour)},
	}

	got := Sequence(trace, "a", "b", "c")

	if !got {
		t.Fatal("expected true")
	}
}

func TestSequence_empty(t *testing.T) {
	t.Parallel()

	base := baseTime()
	trace := Trace{
		{Label: "a", Time: base},
	}

	got := Sequence(trace)

	if !got {
		t.Fatal("expected true for empty events")
	}
}

func TestSequence_emptyTrace(t *testing.T) {
	t.Parallel()

	got := Sequence(Trace{}, "a")

	if got {
		t.Fatal("expected false")
	}
}
