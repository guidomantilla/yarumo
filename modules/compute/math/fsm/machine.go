package fsm

import (
	"cmp"
	"slices"

	"github.com/guidomantilla/yarumo/compute/math/graph"
)

// NewMachine creates a new finite state machine with the given initial state, states, and transitions.
// It validates that the initial state exists, all states are unique, all transitions reference
// existing states, transition IDs are unique, and events are non-empty.
func NewMachine(initial string, states []State, transitions []Transition) (*Machine, error) {
	stateMap := make(map[string]State, len(states))

	for _, s := range states {
		_, exists := stateMap[s.ID]
		if exists {
			return nil, ErrFSM(ErrDuplicateState)
		}

		stateMap[s.ID] = s
	}

	_, exists := stateMap[initial]
	if !exists {
		return nil, ErrFSM(ErrNoInitialState)
	}

	transMap := make(map[string]Transition, len(transitions))
	index := make(map[string][]string)

	for _, tr := range transitions {
		_, exists := transMap[tr.ID]
		if exists {
			return nil, ErrFSM(ErrDuplicateTransition)
		}

		if tr.Event == "" {
			return nil, ErrFSM(ErrInvalidEvent)
		}

		_, fromExists := stateMap[tr.From]
		_, toExists := stateMap[tr.To]

		if !fromExists || !toExists {
			return nil, ErrFSM(ErrInvalidTransition)
		}

		transMap[tr.ID] = tr

		key := tr.From + "\x00" + tr.Event
		index[key] = append(index[key], tr.ID)
	}

	g := graph.NewDirected()

	for _, s := range states {
		_ = g.AddNode(graph.Node{ID: s.ID, Metadata: s.Metadata})
	}

	for _, tr := range transitions {
		_ = g.AddEdge(graph.Edge{ID: tr.ID, From: tr.From, To: tr.To, Label: tr.Event, Metadata: tr.Metadata})
	}

	return &Machine{
		graph:   g,
		states:  stateMap,
		trans:   transMap,
		initial: initial,
		current: initial,
		index:   index,
	}, nil
}

// Current returns the ID of the current state.
func (m *Machine) Current() string {
	return m.current
}

// Send triggers a transition by event name. It evaluates guards in order
// and applies the first matching transition. The ctx parameter is passed to guards.
func (m *Machine) Send(event string, ctx any) error {
	key := m.current + "\x00" + event
	transIDs, exists := m.index[key]

	if !exists || len(transIDs) == 0 {
		return ErrFSM(ErrTransitionNotFound)
	}

	for _, tid := range transIDs {
		tr := m.trans[tid]

		if tr.Guard == nil {
			m.current = tr.To

			return nil
		}

		if tr.Guard(ctx) {
			m.current = tr.To

			return nil
		}
	}

	return ErrFSM(ErrGuardRejected)
}

// Available returns all transitions from the current state, sorted by ID.
// Guards are not evaluated.
func (m *Machine) Available() []Transition {
	result := make([]Transition, 0)

	for _, tr := range m.trans {
		if tr.From == m.current {
			result = append(result, tr)
		}
	}

	slices.SortFunc(result, func(a, b Transition) int {
		return cmp.Compare(a.ID, b.ID)
	})

	return result
}

// Can reports whether the given event can trigger a transition from the current state.
// Guards are evaluated with the provided ctx.
func (m *Machine) Can(event string, ctx any) bool {
	key := m.current + "\x00" + event
	transIDs, exists := m.index[key]

	if !exists || len(transIDs) == 0 {
		return false
	}

	for _, tid := range transIDs {
		tr := m.trans[tid]

		if tr.Guard == nil {
			return true
		}

		if tr.Guard(ctx) {
			return true
		}
	}

	return false
}

// Graph returns a clone of the internal directed graph for analysis.
func (m *Machine) Graph() *graph.Directed {
	return m.graph.CloneDirected()
}

// Reset returns the machine to its initial state.
func (m *Machine) Reset() {
	m.current = m.initial
}

// States returns all states sorted by ID.
func (m *Machine) States() []State {
	result := make([]State, 0, len(m.states))

	for _, s := range m.states {
		result = append(result, s)
	}

	slices.SortFunc(result, func(a, b State) int {
		return cmp.Compare(a.ID, b.ID)
	})

	return result
}

// Transitions returns all transitions sorted by ID.
func (m *Machine) Transitions() []Transition {
	result := make([]Transition, 0, len(m.trans))

	for _, tr := range m.trans {
		result = append(result, tr)
	}

	slices.SortFunc(result, func(a, b Transition) int {
		return cmp.Compare(a.ID, b.ID)
	})

	return result
}

// Alphabet returns all unique event names sorted alphabetically.
func (m *Machine) Alphabet() []string {
	seen := make(map[string]bool)

	for _, tr := range m.trans {
		seen[tr.Event] = true
	}

	result := make([]string, 0, len(seen))

	for event := range seen {
		result = append(result, event)
	}

	slices.Sort(result)

	return result
}

// Reachable returns all state IDs reachable from the given state, sorted alphabetically.
// The given state itself is not included in the result.
func (m *Machine) Reachable(state string) ([]string, error) {
	_, exists := m.states[state]
	if !exists {
		return nil, ErrFSM(ErrStateNotFound)
	}

	result, _ := graph.Reachable(m.graph, state)

	return result, nil
}

// IsReachable reports whether the target state is reachable from the source state.
func (m *Machine) IsReachable(from, to string) (bool, error) {
	_, fromExists := m.states[from]
	if !fromExists {
		return false, ErrFSM(ErrStateNotFound)
	}

	_, toExists := m.states[to]
	if !toExists {
		return false, ErrFSM(ErrStateNotFound)
	}

	reachable, _ := graph.Reachable(m.graph, from)

	return slices.Contains(reachable, to), nil
}

// DeadStates returns state IDs that have no outgoing transitions, sorted alphabetically.
func (m *Machine) DeadStates() []string {
	result := make([]string, 0)

	for id := range m.states {
		outEdges, _ := m.graph.OutEdges(id)

		if len(outEdges) == 0 {
			result = append(result, id)
		}
	}

	slices.Sort(result)

	return result
}

// Deterministic reports whether the machine is deterministic.
// A machine is deterministic if for every (state, event) pair there is at most one transition.
func (m *Machine) Deterministic() bool {
	for _, transIDs := range m.index {
		if len(transIDs) > 1 {
			return false
		}
	}

	return true
}

// IsComplete reports whether every state has at least one transition for every event in the alphabet.
func (m *Machine) IsComplete() bool {
	alphabet := m.Alphabet()

	if len(alphabet) == 0 {
		return true
	}

	for id := range m.states {
		for _, event := range alphabet {
			key := id + "\x00" + event
			transIDs, exists := m.index[key]

			if !exists || len(transIDs) == 0 {
				return false
			}
		}
	}

	return true
}
