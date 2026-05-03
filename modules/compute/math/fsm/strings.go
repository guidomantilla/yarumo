package fsm

import "strings"

// String returns a human-readable representation of a State.
func (s State) String() string {
	return "State(" + s.ID + ")"
}

// String returns a human-readable representation of a Transition.
func (t Transition) String() string {
	var b strings.Builder

	b.WriteString("Transition(")
	b.WriteString(t.ID)
	b.WriteString(": ")
	b.WriteString(t.From)
	b.WriteString(" --[")
	b.WriteString(t.Event)
	b.WriteString("]--> ")
	b.WriteString(t.To)

	if t.Guard != nil {
		b.WriteString(", guarded")
	}

	b.WriteString(")")

	return b.String()
}
