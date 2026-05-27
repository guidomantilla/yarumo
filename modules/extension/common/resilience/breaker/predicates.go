package breaker

// NoopOnStateChange is the default OnStateChangeFn: no-op. Callers that
// want to observe transitions provide their own hook via
// WithOnStateChange.
func NoopOnStateChange(_ State, _ State) {}
