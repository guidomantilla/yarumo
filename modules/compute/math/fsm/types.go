// Package fsm provides finite state machine primitives.
package fsm

import (
	"github.com/guidomantilla/yarumo/compute/math/graph"
)

// Guard evaluates whether a transition is allowed.
// Guards should be pure functions with no side effects.
type Guard func(any) bool

// State represents a state in a finite state machine.
type State struct {
	ID       string
	Metadata any
}

// Transition represents a transition between two states triggered by an event.
type Transition struct {
	ID       string
	From     string // state ID.
	To       string // state ID.
	Event    string
	Guard    Guard // nil means always allowed.
	Metadata any
}

// Machine is a finite state machine backed by a directed graph.
type Machine struct {
	graph   *graph.Directed
	states  map[string]State
	trans   map[string]Transition
	initial string
	current string
	index   map[string][]string // stateID+"\x00"+event → []transitionID.
}
