// Package temporal provides bounded temporal assertions over concrete event traces.
// It checks concrete traces of timestamped events for compliance verification,
// fraud detection, and SLA checking.
package temporal

import "time"

// Event is a labeled timestamped occurrence.
type Event struct {
	Label string
	Time  time.Time
}

// Trace is a chronologically ordered sequence of events.
type Trace []Event

// Violation describes a specific temporal assertion violation.
type Violation struct {
	TriggerEvent Event
	Message      string
}
