package markov

// String returns a human-readable representation of a StateClass.
func (c StateClass) String() string {
	switch c {
	case Transient:
		return "Transient"
	case Recurrent:
		return "Recurrent"
	case Absorbing:
		return "Absorbing"
	default:
		return "Unknown"
	}
}
