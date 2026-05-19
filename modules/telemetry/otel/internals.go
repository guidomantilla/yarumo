package otel

import (
	"context"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// unwindTimeout is the per-provider shutdown budget used when Observe tears
// down already-started providers after a later step fails. Chosen to be
// large enough for an OTLP collector flush but small enough that a startup
// failure does not block boot for long.
const unwindTimeout = 5 * time.Second

// noopStop is the StopFn returned when Observe fails before any provider has
// started successfully, or by each provider's setup function on error. It is
// a no-op that ignores its arguments and never panics.
func noopStop(_ context.Context, _ time.Duration) {}

// resetStatus seeds an atomic.Value with an "optimistic" ExporterStatus:
// Connected=true, no error observed, counters at zero. Called once on
// package init for each provider's status slot.
func resetStatus(v *atomic.Value) {
	v.Store(&ExporterStatus{Connected: true})
}

// loadStatus returns a value-copy snapshot of the ExporterStatus stored in
// v, isolating callers from any subsequent in-place updates.
func loadStatus(v *atomic.Value) ExporterStatus {
	p, ok := v.Load().(*ExporterStatus)
	if !ok || p == nil {
		return ExporterStatus{}
	}
	return *p
}

// recordExport mutates the status stored in v based on the outcome of an
// Export call. On error: Connected=false, LastError/LastErrorTime set, and
// DroppedCount increases by recordCount. On success: Connected=true,
// SuccessCount increases by recordCount. The caller is expected to hold the
// per-exporter mutex so concurrent Export calls do not race.
func recordExport(v *atomic.Value, err error, recordCount int64) {
	prev := loadStatus(v)
	next := prev
	if err != nil {
		next.Connected = false
		next.LastError = err
		next.LastErrorTime = time.Now()
		next.DroppedCount = prev.DroppedCount + recordCount
	} else {
		next.Connected = true
		next.SuccessCount = prev.SuccessCount + recordCount
	}
	v.Store(&next)
}

// countMetricDataPoints sums the data points across every metric in every
// scope of a ResourceMetrics; used as the "records lost" measure when a
// metric Export fails.
func countMetricDataPoints(data *metricdata.ResourceMetrics) int64 {
	if data == nil {
		return 0
	}
	var n int64
	for _, sm := range data.ScopeMetrics {
		n += int64(len(sm.Metrics))
	}
	return n
}
