package engine

// engine is the private implementation of Engine.
type engine struct {
	options Options
}

// NewEngine creates a new inference engine with the given options.
func NewEngine(opts ...Option) Engine {
	return &engine{
		options: NewOptions(opts...),
	}
}
