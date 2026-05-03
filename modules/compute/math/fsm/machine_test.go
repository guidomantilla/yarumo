package fsm

import (
	"errors"
	"testing"
)

func newTrafficLight() (*Machine, error) {
	states := []State{
		{ID: "red"},
		{ID: "yellow"},
		{ID: "green"},
	}

	transitions := []Transition{
		{ID: "t1", From: "red", To: "green", Event: "next"},
		{ID: "t2", From: "green", To: "yellow", Event: "next"},
		{ID: "t3", From: "yellow", To: "red", Event: "next"},
	}

	return NewMachine("red", states, transitions)
}

func TestNewMachine(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Current() != "red" {
		t.Fatalf("expected current state %q, got %q", "red", m.Current())
	}
}

func TestNewMachine_no_initial_state(t *testing.T) {
	t.Parallel()

	_, err := NewMachine("missing", []State{{ID: "a"}}, nil)

	if !errors.Is(err, ErrNoInitialState) {
		t.Fatalf("expected ErrNoInitialState, got %v", err)
	}
}

func TestNewMachine_duplicate_state(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "a"}, {ID: "a"}}
	_, err := NewMachine("a", states, nil)

	if !errors.Is(err, ErrDuplicateState) {
		t.Fatalf("expected ErrDuplicateState, got %v", err)
	}
}

func TestNewMachine_duplicate_transition(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "a"}, {ID: "b"}}
	transitions := []Transition{
		{ID: "t1", From: "a", To: "b", Event: "go"},
		{ID: "t1", From: "b", To: "a", Event: "back"},
	}

	_, err := NewMachine("a", states, transitions)

	if !errors.Is(err, ErrDuplicateTransition) {
		t.Fatalf("expected ErrDuplicateTransition, got %v", err)
	}
}

func TestNewMachine_invalid_event(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "a"}, {ID: "b"}}
	transitions := []Transition{
		{ID: "t1", From: "a", To: "b", Event: ""},
	}

	_, err := NewMachine("a", states, transitions)

	if !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("expected ErrInvalidEvent, got %v", err)
	}
}

func TestNewMachine_invalid_transition_from(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "a"}}
	transitions := []Transition{
		{ID: "t1", From: "missing", To: "a", Event: "go"},
	}

	_, err := NewMachine("a", states, transitions)

	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestNewMachine_invalid_transition_to(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "a"}}
	transitions := []Transition{
		{ID: "t1", From: "a", To: "missing", Event: "go"},
	}

	_, err := NewMachine("a", states, transitions)

	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestNewMachine_no_transitions(t *testing.T) {
	t.Parallel()

	m, err := NewMachine("a", []State{{ID: "a"}}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Current() != "a" {
		t.Fatalf("expected current state %q, got %q", "a", m.Current())
	}
}

func TestNewMachine_with_metadata(t *testing.T) {
	t.Parallel()

	states := []State{
		{ID: "a", Metadata: "state-meta"},
		{ID: "b", Metadata: 42},
	}
	transitions := []Transition{
		{ID: "t1", From: "a", To: "b", Event: "go", Metadata: "trans-meta"},
	}

	m, err := NewMachine("a", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Current() != "a" {
		t.Fatalf("expected current state %q, got %q", "a", m.Current())
	}
}

func TestSend(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = m.Send("next", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Current() != "green" {
		t.Fatalf("expected %q, got %q", "green", m.Current())
	}
}

func TestSend_full_cycle(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = m.Send("next", nil)
	if err != nil {
		t.Fatalf("unexpected error on first send: %v", err)
	}

	err = m.Send("next", nil)
	if err != nil {
		t.Fatalf("unexpected error on second send: %v", err)
	}

	err = m.Send("next", nil)
	if err != nil {
		t.Fatalf("unexpected error on third send: %v", err)
	}

	if m.Current() != "red" {
		t.Fatalf("expected %q after full cycle, got %q", "red", m.Current())
	}
}

func TestSend_no_transition(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = m.Send("invalid", nil)

	if !errors.Is(err, ErrTransitionNotFound) {
		t.Fatalf("expected ErrTransitionNotFound, got %v", err)
	}

	if m.Current() != "red" {
		t.Fatalf("state should not change on failed send, got %q", m.Current())
	}
}

func TestSend_with_guard_pass(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "locked"}, {ID: "unlocked"}}
	transitions := []Transition{
		{
			ID: "t1", From: "locked", To: "unlocked", Event: "unlock",
			Guard: func(ctx any) bool {
				key, ok := ctx.(string)
				return ok && key == "secret"
			},
		},
	}

	m, err := NewMachine("locked", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = m.Send("unlock", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Current() != "unlocked" {
		t.Fatalf("expected %q, got %q", "unlocked", m.Current())
	}
}

func TestSend_with_guard_reject(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "locked"}, {ID: "unlocked"}}
	transitions := []Transition{
		{
			ID: "t1", From: "locked", To: "unlocked", Event: "unlock",
			Guard: func(ctx any) bool {
				key, ok := ctx.(string)
				return ok && key == "secret"
			},
		},
	}

	m, err := NewMachine("locked", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = m.Send("unlock", "wrong")

	if !errors.Is(err, ErrGuardRejected) {
		t.Fatalf("expected ErrGuardRejected, got %v", err)
	}

	if m.Current() != "locked" {
		t.Fatalf("state should not change on rejected guard, got %q", m.Current())
	}
}

func TestSend_multiple_transitions_first_guard_wins(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "idle"}, {ID: "admin"}, {ID: "user"}}
	transitions := []Transition{
		{
			ID: "t1", From: "idle", To: "admin", Event: "login",
			Guard: func(ctx any) bool {
				role, ok := ctx.(string)
				return ok && role == "admin"
			},
		},
		{
			ID: "t2", From: "idle", To: "user", Event: "login",
			Guard: func(ctx any) bool {
				role, ok := ctx.(string)
				return ok && role == "user"
			},
		},
	}

	m, err := NewMachine("idle", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = m.Send("login", "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Current() != "user" {
		t.Fatalf("expected %q, got %q", "user", m.Current())
	}
}

func TestSend_multiple_transitions_all_guards_reject(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "idle"}, {ID: "a"}, {ID: "b"}}
	transitions := []Transition{
		{
			ID: "t1", From: "idle", To: "a", Event: "go",
			Guard: func(any) bool { return false },
		},
		{
			ID: "t2", From: "idle", To: "b", Event: "go",
			Guard: func(any) bool { return false },
		},
	}

	m, err := NewMachine("idle", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = m.Send("go", nil)

	if !errors.Is(err, ErrGuardRejected) {
		t.Fatalf("expected ErrGuardRejected, got %v", err)
	}
}

func TestSend_nil_guard_always_passes(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "a"}, {ID: "b"}}
	transitions := []Transition{
		{ID: "t1", From: "a", To: "b", Event: "go", Guard: nil},
	}

	m, err := NewMachine("a", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = m.Send("go", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Current() != "b" {
		t.Fatalf("expected %q, got %q", "b", m.Current())
	}
}

func TestAvailable(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	avail := m.Available()

	if len(avail) != 1 {
		t.Fatalf("expected 1 available transition, got %d", len(avail))
	}

	if avail[0].ID != "t1" {
		t.Fatalf("expected transition %q, got %q", "t1", avail[0].ID)
	}
}

func TestAvailable_multiple(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "idle"}, {ID: "a"}, {ID: "b"}}
	transitions := []Transition{
		{ID: "t1", From: "idle", To: "a", Event: "go_a"},
		{ID: "t2", From: "idle", To: "b", Event: "go_b"},
	}

	m, err := NewMachine("idle", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	avail := m.Available()

	if len(avail) != 2 {
		t.Fatalf("expected 2 available transitions, got %d", len(avail))
	}

	if avail[0].ID != "t1" {
		t.Fatalf("expected first transition %q, got %q", "t1", avail[0].ID)
	}

	if avail[1].ID != "t2" {
		t.Fatalf("expected second transition %q, got %q", "t2", avail[1].ID)
	}
}

func TestAvailable_no_transitions(t *testing.T) {
	t.Parallel()

	m, err := NewMachine("a", []State{{ID: "a"}}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	avail := m.Available()

	if len(avail) != 0 {
		t.Fatalf("expected 0 available transitions, got %d", len(avail))
	}
}

func TestAvailable_sorted_by_id(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "s"}, {ID: "a"}, {ID: "b"}, {ID: "c"}}
	transitions := []Transition{
		{ID: "z_trans", From: "s", To: "c", Event: "c"},
		{ID: "a_trans", From: "s", To: "a", Event: "a"},
		{ID: "m_trans", From: "s", To: "b", Event: "b"},
	}

	m, err := NewMachine("s", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	avail := m.Available()

	if len(avail) != 3 {
		t.Fatalf("expected 3 available transitions, got %d", len(avail))
	}

	if avail[0].ID != "a_trans" {
		t.Fatalf("expected first %q, got %q", "a_trans", avail[0].ID)
	}

	if avail[1].ID != "m_trans" {
		t.Fatalf("expected second %q, got %q", "m_trans", avail[1].ID)
	}

	if avail[2].ID != "z_trans" {
		t.Fatalf("expected third %q, got %q", "z_trans", avail[2].ID)
	}
}

func TestCan_true(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !m.Can("next", nil) {
		t.Fatal("expected Can to return true for valid event")
	}
}

func TestCan_false_no_transition(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Can("invalid", nil) {
		t.Fatal("expected Can to return false for invalid event")
	}
}

func TestCan_false_guard_rejects(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "a"}, {ID: "b"}}
	transitions := []Transition{
		{
			ID: "t1", From: "a", To: "b", Event: "go",
			Guard: func(any) bool { return false },
		},
	}

	m, err := NewMachine("a", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Can("go", nil) {
		t.Fatal("expected Can to return false when guard rejects")
	}
}

func TestCan_true_guard_passes(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "a"}, {ID: "b"}}
	transitions := []Transition{
		{
			ID: "t1", From: "a", To: "b", Event: "go",
			Guard: func(any) bool { return true },
		},
	}

	m, err := NewMachine("a", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !m.Can("go", "ctx") {
		t.Fatal("expected Can to return true when guard passes")
	}
}

func TestGraph_returns_clone(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := m.Graph()

	if g.NodeCount() != 3 {
		t.Fatalf("expected 3 nodes, got %d", g.NodeCount())
	}

	if g.EdgeCount() != 3 {
		t.Fatalf("expected 3 edges, got %d", g.EdgeCount())
	}

	// Verify it is a clone by modifying it.
	_ = g.RemoveNode("red")

	g2 := m.Graph()

	if g2.NodeCount() != 3 {
		t.Fatalf("original graph should be unaffected, got %d nodes", g2.NodeCount())
	}
}

func TestGraph_preserves_structure(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := m.Graph()

	if !g.HasNode("red") {
		t.Fatal("expected graph to have node 'red'")
	}

	if !g.HasNode("green") {
		t.Fatal("expected graph to have node 'green'")
	}

	if !g.HasNode("yellow") {
		t.Fatal("expected graph to have node 'yellow'")
	}

	if !g.HasEdge("t1") {
		t.Fatal("expected graph to have edge 't1'")
	}
}

func TestReset(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = m.Send("next", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Current() != "green" {
		t.Fatalf("expected %q, got %q", "green", m.Current())
	}

	m.Reset()

	if m.Current() != "red" {
		t.Fatalf("expected %q after reset, got %q", "red", m.Current())
	}
}

func TestReset_idempotent(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m.Reset()

	if m.Current() != "red" {
		t.Fatalf("expected %q, got %q", "red", m.Current())
	}
}

func TestNewMachine_error_is_typed(t *testing.T) {
	t.Parallel()

	_, err := NewMachine("missing", []State{{ID: "a"}}, nil)

	var typed *Error
	if !errors.As(err, &typed) {
		t.Fatal("expected error to be *Error")
	}

	if typed.Type != FSMType {
		t.Fatalf("expected type %q, got %q", FSMType, typed.Type)
	}
}

func TestSend_error_is_typed(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = m.Send("invalid", nil)

	var typed *Error
	if !errors.As(err, &typed) {
		t.Fatal("expected error to be *Error")
	}
}

func TestSend_guard_rejected_error_is_typed(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "a"}, {ID: "b"}}
	transitions := []Transition{
		{
			ID: "t1", From: "a", To: "b", Event: "go",
			Guard: func(any) bool { return false },
		},
	}

	m, err := NewMachine("a", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = m.Send("go", nil)

	var typed *Error
	if !errors.As(err, &typed) {
		t.Fatal("expected error to be *Error")
	}
}

func TestSend_self_loop(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "a"}}
	transitions := []Transition{
		{ID: "t1", From: "a", To: "a", Event: "loop"},
	}

	m, err := NewMachine("a", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = m.Send("loop", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Current() != "a" {
		t.Fatalf("expected %q, got %q", "a", m.Current())
	}
}

func TestAvailable_changes_after_send(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	avail := m.Available()
	if len(avail) != 1 || avail[0].To != "green" {
		t.Fatalf("expected transition to green from red")
	}

	err = m.Send("next", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	avail = m.Available()
	if len(avail) != 1 || avail[0].To != "yellow" {
		t.Fatalf("expected transition to yellow from green")
	}
}

// --- States ---

func TestStates(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	states := m.States()

	if len(states) != 3 {
		t.Fatalf("expected 3 states, got %d", len(states))
	}

	if states[0].ID != "green" {
		t.Fatalf("expected first state %q, got %q", "green", states[0].ID)
	}

	if states[1].ID != "red" {
		t.Fatalf("expected second state %q, got %q", "red", states[1].ID)
	}

	if states[2].ID != "yellow" {
		t.Fatalf("expected third state %q, got %q", "yellow", states[2].ID)
	}
}

func TestStates_single(t *testing.T) {
	t.Parallel()

	m, err := NewMachine("a", []State{{ID: "a"}}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	states := m.States()

	if len(states) != 1 {
		t.Fatalf("expected 1 state, got %d", len(states))
	}
}

// --- Transitions ---

func TestTransitions(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trans := m.Transitions()

	if len(trans) != 3 {
		t.Fatalf("expected 3 transitions, got %d", len(trans))
	}

	if trans[0].ID != "t1" {
		t.Fatalf("expected first transition %q, got %q", "t1", trans[0].ID)
	}

	if trans[1].ID != "t2" {
		t.Fatalf("expected second transition %q, got %q", "t2", trans[1].ID)
	}

	if trans[2].ID != "t3" {
		t.Fatalf("expected third transition %q, got %q", "t3", trans[2].ID)
	}
}

func TestTransitions_empty(t *testing.T) {
	t.Parallel()

	m, err := NewMachine("a", []State{{ID: "a"}}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trans := m.Transitions()

	if len(trans) != 0 {
		t.Fatalf("expected 0 transitions, got %d", len(trans))
	}
}

// --- Alphabet ---

func TestAlphabet(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	alpha := m.Alphabet()

	if len(alpha) != 1 {
		t.Fatalf("expected 1 event, got %d", len(alpha))
	}

	if alpha[0] != "next" {
		t.Fatalf("expected event %q, got %q", "next", alpha[0])
	}
}

func TestAlphabet_multiple_events(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	transitions := []Transition{
		{ID: "t1", From: "a", To: "b", Event: "go"},
		{ID: "t2", From: "b", To: "c", Event: "run"},
		{ID: "t3", From: "c", To: "a", Event: "back"},
	}

	m, err := NewMachine("a", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	alpha := m.Alphabet()

	if len(alpha) != 3 {
		t.Fatalf("expected 3 events, got %d", len(alpha))
	}

	if alpha[0] != "back" {
		t.Fatalf("expected first event %q, got %q", "back", alpha[0])
	}

	if alpha[1] != "go" {
		t.Fatalf("expected second event %q, got %q", "go", alpha[1])
	}

	if alpha[2] != "run" {
		t.Fatalf("expected third event %q, got %q", "run", alpha[2])
	}
}

func TestAlphabet_empty(t *testing.T) {
	t.Parallel()

	m, err := NewMachine("a", []State{{ID: "a"}}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	alpha := m.Alphabet()

	if len(alpha) != 0 {
		t.Fatalf("expected 0 events, got %d", len(alpha))
	}
}

// --- Reachable ---

func TestReachable(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reachable, err := m.Reachable("red")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reachable) != 2 {
		t.Fatalf("expected 2 reachable states, got %d", len(reachable))
	}

	if reachable[0] != "green" {
		t.Fatalf("expected first reachable %q, got %q", "green", reachable[0])
	}

	if reachable[1] != "yellow" {
		t.Fatalf("expected second reachable %q, got %q", "yellow", reachable[1])
	}
}

func TestReachable_isolated(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	transitions := []Transition{
		{ID: "t1", From: "a", To: "b", Event: "go"},
	}

	m, err := NewMachine("a", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reachable, err := m.Reachable("c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(reachable) != 0 {
		t.Fatalf("expected 0 reachable states from isolated state, got %d", len(reachable))
	}
}

func TestReachable_unknown_state(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = m.Reachable("missing")

	if !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("expected ErrStateNotFound, got %v", err)
	}
}

// --- IsReachable ---

func TestIsReachable_true(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ok, err := m.IsReachable("red", "green")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !ok {
		t.Fatal("expected green to be reachable from red")
	}
}

func TestIsReachable_false(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	transitions := []Transition{
		{ID: "t1", From: "a", To: "b", Event: "go"},
	}

	m, err := NewMachine("a", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ok, err := m.IsReachable("c", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ok {
		t.Fatal("expected a to not be reachable from c")
	}
}

func TestIsReachable_unknown_from(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = m.IsReachable("missing", "red")

	if !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("expected ErrStateNotFound, got %v", err)
	}
}

func TestIsReachable_unknown_to(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = m.IsReachable("red", "missing")

	if !errors.Is(err, ErrStateNotFound) {
		t.Fatalf("expected ErrStateNotFound, got %v", err)
	}
}

// --- DeadStates ---

func TestDeadStates(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	transitions := []Transition{
		{ID: "t1", From: "a", To: "b", Event: "go"},
		{ID: "t2", From: "a", To: "c", Event: "run"},
	}

	m, err := NewMachine("a", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dead := m.DeadStates()

	if len(dead) != 2 {
		t.Fatalf("expected 2 dead states, got %d", len(dead))
	}

	if dead[0] != "b" {
		t.Fatalf("expected first dead state %q, got %q", "b", dead[0])
	}

	if dead[1] != "c" {
		t.Fatalf("expected second dead state %q, got %q", "c", dead[1])
	}
}

func TestDeadStates_none(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dead := m.DeadStates()

	if len(dead) != 0 {
		t.Fatalf("expected 0 dead states, got %d", len(dead))
	}
}

func TestDeadStates_all_dead(t *testing.T) {
	t.Parallel()

	m, err := NewMachine("a", []State{{ID: "a"}}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dead := m.DeadStates()

	if len(dead) != 1 {
		t.Fatalf("expected 1 dead state, got %d", len(dead))
	}

	if dead[0] != "a" {
		t.Fatalf("expected dead state %q, got %q", "a", dead[0])
	}
}

// --- Deterministic ---

func TestDeterministic_true(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !m.Deterministic() {
		t.Fatal("expected traffic light to be deterministic")
	}
}

func TestDeterministic_false(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	transitions := []Transition{
		{ID: "t1", From: "a", To: "b", Event: "go"},
		{ID: "t2", From: "a", To: "c", Event: "go"},
	}

	m, err := NewMachine("a", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Deterministic() {
		t.Fatal("expected non-deterministic machine")
	}
}

func TestDeterministic_no_transitions(t *testing.T) {
	t.Parallel()

	m, err := NewMachine("a", []State{{ID: "a"}}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !m.Deterministic() {
		t.Fatal("expected machine with no transitions to be deterministic")
	}
}

// --- IsComplete ---

func TestIsComplete_true(t *testing.T) {
	t.Parallel()

	m, err := newTrafficLight()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !m.IsComplete() {
		t.Fatal("expected traffic light to be complete")
	}
}

func TestIsComplete_false(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	transitions := []Transition{
		{ID: "t1", From: "a", To: "b", Event: "go"},
		{ID: "t2", From: "b", To: "c", Event: "run"},
	}

	m, err := NewMachine("a", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.IsComplete() {
		t.Fatal("expected incomplete machine")
	}
}

func TestIsComplete_no_transitions(t *testing.T) {
	t.Parallel()

	m, err := NewMachine("a", []State{{ID: "a"}}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !m.IsComplete() {
		t.Fatal("expected machine with no transitions to be complete (empty alphabet)")
	}
}

func TestIsComplete_single_event_all_states(t *testing.T) {
	t.Parallel()

	states := []State{{ID: "a"}, {ID: "b"}}
	transitions := []Transition{
		{ID: "t1", From: "a", To: "b", Event: "toggle"},
		{ID: "t2", From: "b", To: "a", Event: "toggle"},
	}

	m, err := NewMachine("a", states, transitions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !m.IsComplete() {
		t.Fatal("expected complete machine")
	}
}
