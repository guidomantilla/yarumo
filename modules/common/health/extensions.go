package health

import "context"

// DefaultHealth is the preconfigured [Health] singleton with default options.
//
// It is intended for application-wide registration of [Check] instances and
// on-demand aggregation via [Status]. Tests and isolated subsystems should
// construct their own [Health] via [NewHealth] to avoid cross-test
// interference.
var DefaultHealth = NewHealth()

// Register adds a [Check] to [DefaultHealth].
func Register(check Check) {
	DefaultHealth.Register(check)
}

// Aggregate delegates to [DefaultHealth.Status]. It is named Aggregate rather
// than Status because Status is already the type name of the enum returned
// here, and Go does not allow a function and a type to share the same name
// in the same package.
func Aggregate(ctx context.Context) (Status, []Result) {
	return DefaultHealth.Status(ctx)
}
