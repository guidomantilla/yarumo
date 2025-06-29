package log

var (
	_ ConfigureFn = Configure
)

type ConfigureFn func(name string, version string, opts ...Option)
