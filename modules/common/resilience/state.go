package resilience

// State represents the operating state of a CircuitBreaker.
type State int

// State values for a CircuitBreaker, mirroring github.com/sony/gobreaker.
const (
	// StateClosed indicates the breaker passes all calls through.
	StateClosed State = iota
	// StateHalfOpen indicates the breaker is probing a limited number of calls.
	StateHalfOpen
	// StateOpen indicates the breaker is failing fast without invoking the call.
	StateOpen
)

// String returns the human-readable name of the state.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return "unknown"
	}
}
