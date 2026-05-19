package otel

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// resetAllStatuses returns the three slots to their package-init state. It
// is used at the start of each Status-related test so prior tests cannot
// leak counters/errors into the current case.
func resetAllStatuses(t *testing.T) {
	t.Helper()
	resetStatus(&tracerStatus)
	resetStatus(&meterStatus)
	resetStatus(&loggerStatus)
}

// fakeTraceExporter is a sdktrace.SpanExporter that returns the configured
// error and counts ExportSpans invocations.
type fakeTraceExporter struct {
	err      error
	calls    atomic.Int64
	shutdown atomic.Bool
}

func (f *fakeTraceExporter) ExportSpans(_ context.Context, _ []sdktrace.ReadOnlySpan) error {
	f.calls.Add(1)
	return f.err
}

func (f *fakeTraceExporter) Shutdown(_ context.Context) error {
	f.shutdown.Store(true)
	return nil
}

// fakeMeterExporter is a minimal sdkmetric.Exporter for tests.
type fakeMeterExporter struct {
	err   error
	calls atomic.Int64
}

func (f *fakeMeterExporter) Temporality(_ sdkmetric.InstrumentKind) metricdata.Temporality {
	return metricdata.CumulativeTemporality
}

func (f *fakeMeterExporter) Aggregation(_ sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	return sdkmetric.AggregationDefault{}
}

func (f *fakeMeterExporter) Export(_ context.Context, _ *metricdata.ResourceMetrics) error {
	f.calls.Add(1)
	return f.err
}

func (f *fakeMeterExporter) ForceFlush(_ context.Context) error { return nil }
func (f *fakeMeterExporter) Shutdown(_ context.Context) error   { return nil }

// fakeLogExporter is a minimal sdklog.Exporter for tests.
type fakeLogExporter struct {
	err   error
	calls atomic.Int64
}

func (f *fakeLogExporter) Export(_ context.Context, _ []sdklog.Record) error {
	f.calls.Add(1)
	return f.err
}

func (f *fakeLogExporter) ForceFlush(_ context.Context) error { return nil }
func (f *fakeLogExporter) Shutdown(_ context.Context) error   { return nil }

func TestTracerStatus_DefaultIsOptimistic(t *testing.T) {
	resetAllStatuses(t)

	got := TracerStatus()
	if !got.Connected {
		t.Fatalf("expected Connected=true on a fresh status")
	}
	if got.LastError != nil {
		t.Fatalf("expected nil LastError, got %v", got.LastError)
	}
	if got.DroppedCount != 0 || got.SuccessCount != 0 {
		t.Fatalf("expected zero counters, got dropped=%d success=%d", got.DroppedCount, got.SuccessCount)
	}
}

func TestRecordingTraceExporter_TracksSuccessAndFailure(t *testing.T) {
	resetAllStatuses(t)

	inner := &fakeTraceExporter{}
	rec := &recordingTraceExporter{inner: inner}

	// 5 spans, success path.
	err := rec.ExportSpans(context.Background(), make([]sdktrace.ReadOnlySpan, 5))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	status := TracerStatus()
	if !status.Connected {
		t.Fatalf("expected Connected=true after success, got %+v", status)
	}
	if status.SuccessCount != 5 {
		t.Fatalf("SuccessCount = %d, want 5", status.SuccessCount)
	}

	// 3 spans, failure path.
	failure := errors.New("collector down")
	inner.err = failure
	err = rec.ExportSpans(context.Background(), make([]sdktrace.ReadOnlySpan, 3))
	if !errors.Is(err, failure) {
		t.Fatalf("inner error not surfaced: %v", err)
	}
	status = TracerStatus()
	if status.Connected {
		t.Fatalf("expected Connected=false after failure")
	}
	if !errors.Is(status.LastError, failure) {
		t.Fatalf("LastError = %v, want %v", status.LastError, failure)
	}
	if status.LastErrorTime.IsZero() {
		t.Fatalf("LastErrorTime must be set after failure")
	}
	if status.DroppedCount != 3 {
		t.Fatalf("DroppedCount = %d, want 3", status.DroppedCount)
	}

	// Recovery: SuccessCount continues to grow, Connected flips back to true.
	inner.err = nil
	err = rec.ExportSpans(context.Background(), make([]sdktrace.ReadOnlySpan, 2))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	status = TracerStatus()
	if !status.Connected {
		t.Fatalf("expected Connected=true on recovery")
	}
	if status.SuccessCount != 7 {
		t.Fatalf("SuccessCount = %d, want 7", status.SuccessCount)
	}
	if status.DroppedCount != 3 {
		t.Fatalf("DroppedCount must not decrease on recovery, got %d", status.DroppedCount)
	}
}

func TestRecordingTraceExporter_ShutdownDelegates(t *testing.T) {
	resetAllStatuses(t)

	inner := &fakeTraceExporter{}
	rec := &recordingTraceExporter{inner: inner}

	err := rec.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !inner.shutdown.Load() {
		t.Fatalf("expected inner.Shutdown to be called")
	}
}

func TestRecordingMeterExporter_TracksExportOutcome(t *testing.T) {
	resetAllStatuses(t)

	data := &metricdata.ResourceMetrics{
		ScopeMetrics: []metricdata.ScopeMetrics{
			{Metrics: []metricdata.Metrics{{}, {}}},
			{Metrics: []metricdata.Metrics{{}}},
		},
	}

	inner := &fakeMeterExporter{}
	rec := &recordingMeterExporter{inner: inner}

	// Success: SuccessCount += 3 (2 + 1 metrics across scopes).
	err := rec.Export(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	status := MeterStatus()
	if !status.Connected || status.SuccessCount != 3 {
		t.Fatalf("after success want Connected=true SuccessCount=3, got %+v", status)
	}

	// Failure: DroppedCount += 3.
	failure := errors.New("collector down")
	inner.err = failure
	err = rec.Export(context.Background(), data)
	if !errors.Is(err, failure) {
		t.Fatalf("inner error not surfaced: %v", err)
	}
	status = MeterStatus()
	if status.Connected || status.DroppedCount != 3 {
		t.Fatalf("after failure want Connected=false DroppedCount=3, got %+v", status)
	}
}

func TestRecordingLogExporter_TracksExportOutcome(t *testing.T) {
	resetAllStatuses(t)

	inner := &fakeLogExporter{}
	rec := &recordingLogExporter{inner: inner}

	// Success.
	err := rec.Export(context.Background(), make([]sdklog.Record, 4))
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	status := LoggerStatus()
	if !status.Connected || status.SuccessCount != 4 {
		t.Fatalf("after success want Connected=true SuccessCount=4, got %+v", status)
	}

	// Failure.
	failure := errors.New("logger backend down")
	inner.err = failure
	err = rec.Export(context.Background(), make([]sdklog.Record, 2))
	if !errors.Is(err, failure) {
		t.Fatalf("inner error not surfaced: %v", err)
	}
	status = LoggerStatus()
	if status.Connected || status.DroppedCount != 2 || !errors.Is(status.LastError, failure) {
		t.Fatalf("after failure want Connected=false DroppedCount=2 LastError=failure, got %+v", status)
	}
}

func TestObserveStatus_AggregatesAllThree(t *testing.T) {
	resetAllStatuses(t)

	// Seed each slot deterministically.
	tracerStatus.Store(&ExporterStatus{Connected: true, SuccessCount: 10})
	meterStatus.Store(&ExporterStatus{Connected: false, DroppedCount: 5})
	loggerStatus.Store(&ExporterStatus{Connected: true, SuccessCount: 3, DroppedCount: 1})

	tr, mt, lg := ObserveStatus()
	if !tr.Connected || tr.SuccessCount != 10 {
		t.Fatalf("tracer: got %+v", tr)
	}
	if mt.Connected || mt.DroppedCount != 5 {
		t.Fatalf("meter: got %+v", mt)
	}
	if !lg.Connected || lg.SuccessCount != 3 || lg.DroppedCount != 1 {
		t.Fatalf("logger: got %+v", lg)
	}
}

func TestRecordingTraceExporter_ConcurrentExportsAreSafe(t *testing.T) {
	resetAllStatuses(t)

	inner := &fakeTraceExporter{}
	rec := &recordingTraceExporter{inner: inner}

	const workers = 16
	const calls = 100
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < calls; j++ {
				_ = rec.ExportSpans(context.Background(), make([]sdktrace.ReadOnlySpan, 1))
			}
		}()
	}
	wg.Wait()

	status := TracerStatus()
	if status.SuccessCount != int64(workers*calls) {
		t.Fatalf("SuccessCount = %d, want %d", status.SuccessCount, workers*calls)
	}
}

func TestCountMetricDataPoints(t *testing.T) {
	t.Parallel()

	t.Run("nil data returns zero", func(t *testing.T) {
		t.Parallel()
		if got := countMetricDataPoints(nil); got != 0 {
			t.Fatalf("got %d, want 0", got)
		}
	})

	t.Run("sums across scopes", func(t *testing.T) {
		t.Parallel()
		data := &metricdata.ResourceMetrics{
			ScopeMetrics: []metricdata.ScopeMetrics{
				{Metrics: []metricdata.Metrics{{}, {}, {}}},
				{Metrics: []metricdata.Metrics{{}}},
				{Metrics: []metricdata.Metrics{}},
			},
		}
		if got := countMetricDataPoints(data); got != 4 {
			t.Fatalf("got %d, want 4", got)
		}
	})
}
