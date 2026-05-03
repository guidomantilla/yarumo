package evaluate

import (
	"context"
	"testing"
	"time"
)

func TestEntry_Fields(t *testing.T) {
	t.Parallel()

	entry := Entry{
		ID:             "test-id",
		Timestamp:      time.Now(),
		RuleSetName:    "test",
		RuleSetVersion: "1",
		Paradigm:       "deductive",
		Request:        "req",
		Result:         "res",
		Explanation:    "exp",
		Duration:       time.Second,
	}

	if entry.ID != "test-id" {
		t.Fatalf("expected test-id, got %s", entry.ID)
	}

	if entry.RuleSetName != "test" {
		t.Fatalf("expected test, got %s", entry.RuleSetName)
	}
}

type stubLog struct {
	entries []Entry
}

func (s *stubLog) Record(_ context.Context, entry Entry) error {
	s.entries = append(s.entries, entry)

	return nil
}

// Verify interface compliance.
var _ Log = (*stubLog)(nil)

func TestLog_Interface(t *testing.T) {
	t.Parallel()

	l := &stubLog{}
	err := l.Record(context.Background(), Entry{ID: "1"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(l.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(l.entries))
	}
}
