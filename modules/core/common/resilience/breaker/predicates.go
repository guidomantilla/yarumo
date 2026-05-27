package breaker

// NoopOnStateChange is the default OnStateChangeFn: no-op. Callers that
// want to observe transitions provide their own hook via
// WithOnStateChange on the concrete Breaker implementation.
func NoopOnStateChange(_ State, _ State) {}
