package cache

import (
	"context"
	"sync"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// otelCounters holds the four cache counters that get incremented per cache event.
type otelCounters struct {
	hits      metric.Int64Counter
	misses    metric.Int64Counter
	sets      metric.Int64Counter
	evictions metric.Int64Counter
}

// otelMetrics is the lazily-initialized OTel metrics adapter for a cache instance.
type otelMetrics struct {
	once     sync.Once
	initErr  error
	counters otelCounters
	meter    metric.Meter
}

// newOtelMetrics constructs an uninitialized adapter pointing at the given meter name.
func newOtelMetrics(meterName string) *otelMetrics {
	return &otelMetrics{
		meter: otel.GetMeterProvider().Meter(meterName),
	}
}

// init lazily creates the four counters on the meter provider configured at
// construction time. The first call seeds initErr; subsequent calls reuse it.
func (m *otelMetrics) init() error {
	m.once.Do(func() {
		counters, err := buildCounters(m.meter)
		if err != nil {
			m.initErr = cerrs.Wrap(err)
			return
		}
		m.counters = counters
	})

	return m.initErr
}

// buildCounters creates the four cache counters on the given meter. It exists
// as a separate helper so the init path stays compact and test-friendly.
func buildCounters(meter metric.Meter) (otelCounters, error) {
	hits, err := meter.Int64Counter(MetricHits, metric.WithDescription("Cache hits"))
	if err != nil {
		return otelCounters{}, cerrs.Wrap(err)
	}
	misses, err := meter.Int64Counter(MetricMisses, metric.WithDescription("Cache misses"))
	if err != nil {
		return otelCounters{}, cerrs.Wrap(err)
	}
	sets, err := meter.Int64Counter(MetricSets, metric.WithDescription("Cache writes"))
	if err != nil {
		return otelCounters{}, cerrs.Wrap(err)
	}
	evictions, err := meter.Int64Counter(MetricEvictions, metric.WithDescription("Cache evictions"))
	if err != nil {
		return otelCounters{}, cerrs.Wrap(err)
	}
	return otelCounters{hits: hits, misses: misses, sets: sets, evictions: evictions}, nil
}

// recordHit increments the cache.hits counter. It silently ignores
// initialization errors because telemetry must never disrupt the data path.
func (m *otelMetrics) recordHit(ctx context.Context) {
	err := m.init()
	if err != nil {
		return
	}
	m.counters.hits.Add(ctx, 1)
}

// recordMiss increments the cache.misses counter.
func (m *otelMetrics) recordMiss(ctx context.Context) {
	err := m.init()
	if err != nil {
		return
	}
	m.counters.misses.Add(ctx, 1)
}

// recordSet increments the cache.sets counter.
func (m *otelMetrics) recordSet(ctx context.Context) {
	err := m.init()
	if err != nil {
		return
	}
	m.counters.sets.Add(ctx, 1)
}

// recordEviction increments the cache.evictions counter.
func (m *otelMetrics) recordEviction(ctx context.Context) {
	err := m.init()
	if err != nil {
		return
	}
	m.counters.evictions.Add(ctx, 1)
}
