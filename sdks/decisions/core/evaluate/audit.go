package evaluate

import (
	"context"
	"time"
)

// Entry holds the complete context of a decision for audit purposes.
type Entry struct {
	// ID is a unique identifier for this audit entry.
	ID string
	// Timestamp records when the decision was made.
	Timestamp time.Time
	// RuleSetName identifies the ruleset used.
	RuleSetName string
	// RuleSetVersion identifies the ruleset version used.
	RuleSetVersion string
	// Paradigm identifies the reasoning paradigm used.
	Paradigm string
	// Request is the original decision request (serialized).
	Request any
	// Result is the decision result (serialized).
	Result any
	// Explanation is the human-readable explanation.
	Explanation string
	// Duration records how long the decision took.
	Duration time.Duration
}

// Log defines the interface for persisting decision audit entries.
// Implementations are provided by the consuming application (e.g., PostgresAuditLog).
type Log interface {
	// Record persists a decision audit entry.
	Record(ctx context.Context, entry Entry) error
}
