package health

// Status represents the health classification of a probe or an aggregate.
//
// Ordering, from worst to best:
// [StatusUnhealthy] > [StatusDegraded] > [StatusHealthy] > [StatusUnknown].
type Status int

// Status values, ordered so that a higher integer means a worse status.
// This ordering is what enables the worst-status-wins aggregation rule.
const (
	// StatusUnknown indicates the probe has no information yet or no checks are registered.
	StatusUnknown Status = iota
	// StatusHealthy indicates the probe passed without issues.
	StatusHealthy
	// StatusDegraded indicates the probe completed with non-fatal warnings.
	StatusDegraded
	// StatusUnhealthy indicates the probe failed.
	StatusUnhealthy
)

// String returns the textual representation of the status.
func (s Status) String() string {
	switch s {
	case StatusUnknown:
		return "unknown"
	case StatusHealthy:
		return "healthy"
	case StatusDegraded:
		return "degraded"
	case StatusUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}
