package otel

import (
	"context"
	"sync"
	"sync/atomic"

	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Per-provider status snapshots. Each Tracer/Meter/Logger wraps its OTLP
// exporter in a recording*Exporter that mutates the matching atomic.Value
// on every Export call. Status accessors read the latest snapshot.
var (
	tracerStatus atomic.Value // *ExporterStatus
	meterStatus  atomic.Value // *ExporterStatus
	loggerStatus atomic.Value // *ExporterStatus
)

func init() {
	resetStatus(&tracerStatus)
	resetStatus(&meterStatus)
	resetStatus(&loggerStatus)
}

// TracerStatus returns a snapshot of the tracer exporter's health.
func TracerStatus() ExporterStatus {
	return loadStatus(&tracerStatus)
}

// MeterStatus returns a snapshot of the meter exporter's health.
func MeterStatus() ExporterStatus {
	return loadStatus(&meterStatus)
}

// LoggerStatus returns a snapshot of the logger exporter's health.
func LoggerStatus() ExporterStatus {
	return loadStatus(&loggerStatus)
}

// ObserveStatus returns the three per-provider snapshots in one call. The
// snapshots are taken independently — they are not guaranteed to be
// consistent with one another in time.
func ObserveStatus() (tracer ExporterStatus, meter ExporterStatus, logger ExporterStatus) {
	return TracerStatus(), MeterStatus(), LoggerStatus()
}

// recordingTraceExporter wraps sdktrace.SpanExporter, forwarding every call
// while updating tracerStatus on success / failure.
type recordingTraceExporter struct {
	inner sdktrace.SpanExporter
	mu    sync.Mutex
}

// ExportSpans delegates to the inner exporter, recording success or failure
// in tracerStatus. The dropped count grows by len(spans) on failure.
func (e *recordingTraceExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	err := e.inner.ExportSpans(ctx, spans)
	e.mu.Lock()
	defer e.mu.Unlock()
	recordExport(&tracerStatus, err, int64(len(spans)))
	return err
}

// Shutdown delegates to the inner exporter.
func (e *recordingTraceExporter) Shutdown(ctx context.Context) error {
	return e.inner.Shutdown(ctx)
}

// recordingMeterExporter wraps sdkmetric.Exporter, forwarding every call
// while updating meterStatus on success / failure of Export.
type recordingMeterExporter struct {
	inner sdkmetric.Exporter
	mu    sync.Mutex
}

// Temporality delegates to the inner exporter.
func (e *recordingMeterExporter) Temporality(k sdkmetric.InstrumentKind) metricdata.Temporality {
	return e.inner.Temporality(k)
}

// Aggregation delegates to the inner exporter.
func (e *recordingMeterExporter) Aggregation(k sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	return e.inner.Aggregation(k)
}

// Export delegates to the inner exporter, recording success or failure in
// meterStatus. The dropped count grows by the number of data points lost
// on failure (counted across scopes).
func (e *recordingMeterExporter) Export(ctx context.Context, data *metricdata.ResourceMetrics) error {
	err := e.inner.Export(ctx, data)
	e.mu.Lock()
	defer e.mu.Unlock()
	recordExport(&meterStatus, err, countMetricDataPoints(data))
	return err
}

// ForceFlush delegates to the inner exporter.
func (e *recordingMeterExporter) ForceFlush(ctx context.Context) error {
	return e.inner.ForceFlush(ctx)
}

// Shutdown delegates to the inner exporter.
func (e *recordingMeterExporter) Shutdown(ctx context.Context) error {
	return e.inner.Shutdown(ctx)
}

// recordingLogExporter wraps sdklog.Exporter, forwarding every call while
// updating loggerStatus on success / failure of Export.
type recordingLogExporter struct {
	inner sdklog.Exporter
	mu    sync.Mutex
}

// Export delegates to the inner exporter, recording success or failure in
// loggerStatus. The dropped count grows by len(records) on failure.
func (e *recordingLogExporter) Export(ctx context.Context, records []sdklog.Record) error {
	err := e.inner.Export(ctx, records)
	e.mu.Lock()
	defer e.mu.Unlock()
	recordExport(&loggerStatus, err, int64(len(records)))
	return err
}

// ForceFlush delegates to the inner exporter.
func (e *recordingLogExporter) ForceFlush(ctx context.Context) error {
	return e.inner.ForceFlush(ctx)
}

// Shutdown delegates to the inner exporter.
func (e *recordingLogExporter) Shutdown(ctx context.Context) error {
	return e.inner.Shutdown(ctx)
}
