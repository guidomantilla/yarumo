package temporal

import (
	"fmt"
	"time"
)

// ResponseWithin checks that every occurrence of trigger is followed by response within maxDuration.
// Returns true if all triggers have matching responses, and a list of violations for those that do not.
func ResponseWithin(trace Trace, trigger, response string, maxDuration time.Duration) (bool, []Violation) {
	var violations []Violation

	for _, ev := range trace {
		if ev.Label != trigger {
			continue
		}

		deadline := ev.Time.Add(maxDuration)
		found := false

		for _, candidate := range trace {
			if candidate.Label == response && candidate.Time.After(ev.Time) && !candidate.Time.After(deadline) {
				found = true
				break
			}
		}

		if !found {
			violations = append(violations, Violation{
				TriggerEvent: ev,
				Message:      fmt.Sprintf("%s not followed by %s within %s", trigger, response, maxDuration),
			})
		}
	}

	return len(violations) == 0, violations
}

// FrequencyWithin checks if event occurs at least minCount times in any sliding window of windowDur.
// Returns true if such a window is found, which is useful for fraud and anomaly detection.
func FrequencyWithin(trace Trace, event string, minCount int, windowDur time.Duration) bool {
	if minCount < 1 {
		return true
	}

	var times []time.Time

	for _, ev := range trace {
		if ev.Label == event {
			times = append(times, ev.Time)
		}
	}

	if len(times) < minCount {
		return false
	}

	for i := 0; i <= len(times)-minCount; i++ {
		windowEnd := times[i].Add(windowDur)
		count := 0

		for j := i; j < len(times); j++ {
			if !times[j].After(windowEnd) {
				count++
			} else {
				break
			}
		}

		if count >= minCount {
			return true
		}
	}

	return false
}

// Eventually checks if the given event label appears at least once in the trace.
func Eventually(trace Trace, event string) bool {
	for _, ev := range trace {
		if ev.Label == event {
			return true
		}
	}

	return false
}

// Before checks that every occurrence of a appears before the first occurrence of b.
// Returns true if no a appears after the first b. Returns true if b never appears.
func Before(trace Trace, a, b string) bool {
	firstB := time.Time{}
	found := false

	for _, ev := range trace {
		if ev.Label == b {
			firstB = ev.Time
			found = true

			break
		}
	}

	if !found {
		return true
	}

	for _, ev := range trace {
		if ev.Label == a && !ev.Time.Before(firstB) {
			return false
		}
	}

	return true
}

// Elapsed returns the duration between the first occurrence of from and the first occurrence of to.
// Returns ErrEventNotFound if either event is not present.
func Elapsed(trace Trace, from, to string) (time.Duration, error) {
	var fromTime, toTime time.Time

	foundFrom, foundTo := false, false

	for _, ev := range trace {
		if ev.Label == from && !foundFrom {
			fromTime = ev.Time
			foundFrom = true
		}

		if ev.Label == to && !foundTo {
			toTime = ev.Time
			foundTo = true
		}
	}

	if !foundFrom || !foundTo {
		return 0, ErrEventNotFound
	}

	return toTime.Sub(fromTime), nil
}

// Always checks that the predicate holds for every event in the trace.
// Returns true for an empty trace (vacuous truth).
func Always(trace Trace, predicate func(Event) bool) bool {
	for _, ev := range trace {
		if !predicate(ev) {
			return false
		}
	}

	return true
}

// Next checks that the event immediately after the first occurrence of event satisfies the predicate.
// Returns false if event is not found or is the last element in the trace.
func Next(trace Trace, event string, predicate func(Event) bool) bool {
	for i, ev := range trace {
		if ev.Label == event {
			if i+1 >= len(trace) {
				return false
			}

			return predicate(trace[i+1])
		}
	}

	return false
}

// Until checks that a occurs at every position until b occurs, and b must eventually occur.
// Implements the standard LTL Until operator over finite traces.
func Until(trace Trace, a, b string) bool {
	for _, ev := range trace {
		if ev.Label == b {
			return true
		}

		if ev.Label != a {
			return false
		}
	}

	return false
}

// Release checks that b holds at every position until a and b hold together, or b holds forever.
// Implements the standard LTL Release operator, the dual of Until.
func Release(trace Trace, a, b string) bool {
	for _, ev := range trace {
		if ev.Label != b {
			return false
		}

		if ev.Label == a {
			return true
		}
	}

	return true
}

// Since checks that a holds at every position going backward until b is found.
// Implements the past-time LTL Since operator over finite traces.
func Since(trace Trace, a, b string) bool {
	n := len(trace)

	for i := n - 1; i >= 0; i-- {
		ev := trace[i]

		if ev.Label == b {
			return true
		}

		if ev.Label != a {
			return false
		}
	}

	return false
}

// Sequence checks that the given event labels appear in order in the trace, not necessarily consecutive.
// Returns true if a subsequence matching the given order exists.
func Sequence(trace Trace, events ...string) bool {
	if len(events) == 0 {
		return true
	}

	idx := 0

	for _, ev := range trace {
		if ev.Label == events[idx] {
			idx++

			if idx == len(events) {
				return true
			}
		}
	}

	return false
}
