package cache

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/embedded"
)

func TestNewOtelMetrics(t *testing.T) {
	t.Parallel()

	m := newOtelMetrics("cache-test")
	if m == nil {
		t.Fatal("expected non-nil metrics adapter")
	}
	if m.meter == nil {
		t.Fatal("expected meter to be set")
	}
}

func TestOtelMetrics_RecordWithInitErr(t *testing.T) {
	t.Parallel()

	// When init has failed, record* methods must early-return without panicking.
	// We pre-seed initErr by closing once with a stub error and triggering Once.
	m := &otelMetrics{initErr: errors.New("init failed")}
	m.once.Do(func() {})

	// All four record* methods must silently no-op.
	m.recordHit(context.Background())
	m.recordMiss(context.Background())
	m.recordSet(context.Background())
	m.recordEviction(context.Background())
}

func TestOtelMetrics_InitIdempotent(t *testing.T) {
	t.Parallel()

	m := newOtelMetrics("cache-test")
	err1 := m.init()
	if err1 != nil {
		t.Fatalf("unexpected init error: %v", err1)
	}
	err2 := m.init()
	if err2 != nil {
		t.Fatalf("second init should also succeed, got %v", err2)
	}
}

// failingMeter is a metric.Meter stub that fails every counter creation; it
// drives the init/build-counters error paths in metrics.go from tests.
type failingMeter struct {
	embedded.Meter
}

var errStubCounter = errors.New("stub counter failure")

func (failingMeter) Int64Counter(_ string, _ ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	return nil, errStubCounter
}

func (failingMeter) Int64UpDownCounter(_ string, _ ...metric.Int64UpDownCounterOption) (metric.Int64UpDownCounter, error) {
	return nil, errStubCounter
}

func (failingMeter) Int64Histogram(_ string, _ ...metric.Int64HistogramOption) (metric.Int64Histogram, error) {
	return nil, errStubCounter
}

func (failingMeter) Int64Gauge(_ string, _ ...metric.Int64GaugeOption) (metric.Int64Gauge, error) {
	return nil, errStubCounter
}

func (failingMeter) Int64ObservableCounter(_ string, _ ...metric.Int64ObservableCounterOption) (metric.Int64ObservableCounter, error) {
	return nil, errStubCounter
}

func (failingMeter) Int64ObservableUpDownCounter(_ string, _ ...metric.Int64ObservableUpDownCounterOption) (metric.Int64ObservableUpDownCounter, error) {
	return nil, errStubCounter
}

func (failingMeter) Int64ObservableGauge(_ string, _ ...metric.Int64ObservableGaugeOption) (metric.Int64ObservableGauge, error) {
	return nil, errStubCounter
}

func (failingMeter) Float64Counter(_ string, _ ...metric.Float64CounterOption) (metric.Float64Counter, error) {
	return nil, errStubCounter
}

func (failingMeter) Float64UpDownCounter(_ string, _ ...metric.Float64UpDownCounterOption) (metric.Float64UpDownCounter, error) {
	return nil, errStubCounter
}

func (failingMeter) Float64Histogram(_ string, _ ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
	return nil, errStubCounter
}

func (failingMeter) Float64Gauge(_ string, _ ...metric.Float64GaugeOption) (metric.Float64Gauge, error) {
	return nil, errStubCounter
}

func (failingMeter) Float64ObservableCounter(_ string, _ ...metric.Float64ObservableCounterOption) (metric.Float64ObservableCounter, error) {
	return nil, errStubCounter
}

func (failingMeter) Float64ObservableUpDownCounter(_ string, _ ...metric.Float64ObservableUpDownCounterOption) (metric.Float64ObservableUpDownCounter, error) {
	return nil, errStubCounter
}

func (failingMeter) Float64ObservableGauge(_ string, _ ...metric.Float64ObservableGaugeOption) (metric.Float64ObservableGauge, error) {
	return nil, errStubCounter
}

func (failingMeter) RegisterCallback(_ metric.Callback, _ ...metric.Observable) (metric.Registration, error) {
	return nil, errStubCounter
}

func TestBuildCounters_PropagatesError(t *testing.T) {
	t.Parallel()

	_, err := buildCounters(failingMeter{})
	if err == nil {
		t.Fatal("expected error from failing meter")
	}
	if !errors.Is(err, errStubCounter) {
		t.Fatalf("expected wrapped stub error, got %v", err)
	}
}

func TestOtelMetrics_InitPropagatesError(t *testing.T) {
	t.Parallel()

	m := &otelMetrics{meter: failingMeter{}}
	err := m.init()
	if err == nil {
		t.Fatal("expected init error from failing meter")
	}
	if !errors.Is(err, errStubCounter) {
		t.Fatalf("expected wrapped stub error, got %v", err)
	}
}
