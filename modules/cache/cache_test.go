package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

const (
	testValueString = "value"
)

func TestNewCache(t *testing.T) {
	t.Parallel()

	t.Run("ristretto backend", func(t *testing.T) {
		t.Parallel()

		c, err := NewCache[string, []byte](WithBackend(BackendRistretto))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = c.Stop(context.Background()) }()
		if c == nil {
			t.Fatal("expected non-nil cache")
		}
	})

	t.Run("bigcache backend", func(t *testing.T) {
		t.Parallel()

		c, err := NewCache[string, []byte](WithBackend(BackendBigcache))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = c.Stop(context.Background()) }()
		if c == nil {
			t.Fatal("expected non-nil cache")
		}
	})

	t.Run("go-cache backend", func(t *testing.T) {
		t.Parallel()

		c, err := NewCache[string, []byte](WithBackend(BackendGoCache))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = c.Stop(context.Background()) }()
		if c == nil {
			t.Fatal("expected non-nil cache")
		}
	})

	t.Run("default backend is ristretto", func(t *testing.T) {
		t.Parallel()

		c, err := NewCache[string, []byte]()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = c.Stop(context.Background()) }()
		if c == nil {
			t.Fatal("expected non-nil cache")
		}
	})
}

func TestCache_SetGet_Ristretto(t *testing.T) {
	t.Parallel()

	c, err := NewCache[string, []byte](WithBackend(BackendRistretto))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = c.Stop(context.Background()) }()

	ctx := context.Background()
	setErr := c.Set(ctx, "key", []byte(testValueString), time.Minute)
	if setErr != nil {
		t.Fatalf("set failed: %v", setErr)
	}

	got, getErr := c.Get(ctx, "key")
	if getErr != nil {
		t.Fatalf("get failed: %v", getErr)
	}
	if string(got) != testValueString {
		t.Fatalf("got %q, want value", string(got))
	}
}

func TestCache_SetGet_Bigcache(t *testing.T) {
	t.Parallel()

	c, err := NewCache[string, []byte](WithBackend(BackendBigcache))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = c.Stop(context.Background()) }()

	ctx := context.Background()
	setErr := c.Set(ctx, "key", []byte(testValueString), time.Minute)
	if setErr != nil {
		t.Fatalf("set failed: %v", setErr)
	}

	got, getErr := c.Get(ctx, "key")
	if getErr != nil {
		t.Fatalf("get failed: %v", getErr)
	}
	if string(got) != testValueString {
		t.Fatalf("got %q, want value", string(got))
	}
}

func TestCache_SetGet_GoCache(t *testing.T) {
	t.Parallel()

	c, err := NewCache[string, []byte](WithBackend(BackendGoCache))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = c.Stop(context.Background()) }()

	ctx := context.Background()
	setErr := c.Set(ctx, "key", []byte(testValueString), time.Minute)
	if setErr != nil {
		t.Fatalf("set failed: %v", setErr)
	}

	got, getErr := c.Get(ctx, "key")
	if getErr != nil {
		t.Fatalf("get failed: %v", getErr)
	}
	if string(got) != testValueString {
		t.Fatalf("got %q, want value", string(got))
	}
}

func TestCache_Miss(t *testing.T) {
	t.Parallel()

	t.Run("go-cache miss", func(t *testing.T) {
		t.Parallel()

		c, err := NewCache[string, []byte](WithBackend(BackendGoCache))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = c.Stop(context.Background()) }()

		_, getErr := c.Get(context.Background(), "absent")
		if !errors.Is(getErr, ErrCacheMiss) {
			t.Fatalf("expected ErrCacheMiss, got %v", getErr)
		}
	})

	t.Run("ristretto miss", func(t *testing.T) {
		t.Parallel()

		c, err := NewCache[string, []byte](WithBackend(BackendRistretto))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = c.Stop(context.Background()) }()

		_, getErr := c.Get(context.Background(), "absent")
		if !errors.Is(getErr, ErrCacheMiss) {
			t.Fatalf("expected ErrCacheMiss, got %v", getErr)
		}
	})

	t.Run("bigcache miss", func(t *testing.T) {
		t.Parallel()

		c, err := NewCache[string, []byte](WithBackend(BackendBigcache))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = c.Stop(context.Background()) }()

		_, getErr := c.Get(context.Background(), "absent")
		if !errors.Is(getErr, ErrCacheMiss) {
			t.Fatalf("expected ErrCacheMiss, got %v", getErr)
		}
	})
}

func TestCache_Delete(t *testing.T) {
	t.Parallel()

	runDelete := func(t *testing.T, backend Backend) {
		t.Helper()
		c, err := NewCache[string, []byte](WithBackend(backend))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = c.Stop(context.Background()) }()

		ctx := context.Background()
		setErr := c.Set(ctx, "key", []byte(testValueString), time.Minute)
		if setErr != nil {
			t.Fatalf("set failed: %v", setErr)
		}

		delErr := c.Delete(ctx, "key")
		if delErr != nil {
			t.Fatalf("delete failed: %v", delErr)
		}

		_, getErr := c.Get(ctx, "key")
		if !errors.Is(getErr, ErrCacheMiss) {
			t.Fatalf("expected miss after delete, got %v", getErr)
		}
	}

	t.Run("go-cache", func(t *testing.T) {
		t.Parallel()
		runDelete(t, BackendGoCache)
	})
	t.Run("ristretto", func(t *testing.T) {
		t.Parallel()
		runDelete(t, BackendRistretto)
	})
	t.Run("bigcache", func(t *testing.T) {
		t.Parallel()
		runDelete(t, BackendBigcache)
	})
}

func TestCache_Has(t *testing.T) {
	t.Parallel()

	runHas := func(t *testing.T, backend Backend) {
		t.Helper()
		c, err := NewCache[string, []byte](WithBackend(backend))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = c.Stop(context.Background()) }()

		ctx := context.Background()
		if c.Has(ctx, "absent") {
			t.Fatal("expected Has=false for absent key")
		}

		setErr := c.Set(ctx, "present", []byte(testValueString), time.Minute)
		if setErr != nil {
			t.Fatalf("set failed: %v", setErr)
		}

		if !c.Has(ctx, "present") {
			t.Fatal("expected Has=true after set")
		}
	}

	t.Run("go-cache", func(t *testing.T) {
		t.Parallel()
		runHas(t, BackendGoCache)
	})
	t.Run("ristretto", func(t *testing.T) {
		t.Parallel()
		runHas(t, BackendRistretto)
	})
	t.Run("bigcache", func(t *testing.T) {
		t.Parallel()
		runHas(t, BackendBigcache)
	})
}

func TestCache_Clear(t *testing.T) {
	t.Parallel()

	runClear := func(t *testing.T, backend Backend) {
		t.Helper()
		c, err := NewCache[string, []byte](WithBackend(backend))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = c.Stop(context.Background()) }()

		ctx := context.Background()
		setErr := c.Set(ctx, "a", []byte("1"), time.Minute)
		if setErr != nil {
			t.Fatalf("set failed: %v", setErr)
		}
		setErr = c.Set(ctx, "b", []byte("2"), time.Minute)
		if setErr != nil {
			t.Fatalf("set failed: %v", setErr)
		}

		clearErr := c.Clear(ctx)
		if clearErr != nil {
			t.Fatalf("clear failed: %v", clearErr)
		}

		_, getErr := c.Get(ctx, "a")
		if !errors.Is(getErr, ErrCacheMiss) {
			t.Fatalf("expected miss after clear, got %v", getErr)
		}
	}

	t.Run("go-cache", func(t *testing.T) {
		t.Parallel()
		runClear(t, BackendGoCache)
	})
	t.Run("ristretto", func(t *testing.T) {
		t.Parallel()
		runClear(t, BackendRistretto)
	})
	t.Run("bigcache", func(t *testing.T) {
		t.Parallel()
		runClear(t, BackendBigcache)
	})
}

func TestCache_TTLExpiry(t *testing.T) {
	t.Parallel()

	t.Run("go-cache", func(t *testing.T) {
		t.Parallel()

		c, err := NewCache[string, []byte](WithBackend(BackendGoCache), WithGoCacheCapacity(100*time.Millisecond, 50*time.Millisecond))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = c.Stop(context.Background()) }()

		ctx := context.Background()
		setErr := c.Set(ctx, "tempkey", []byte("v"), 50*time.Millisecond)
		if setErr != nil {
			t.Fatalf("set failed: %v", setErr)
		}

		time.Sleep(200 * time.Millisecond)

		_, getErr := c.Get(ctx, "tempkey")
		if !errors.Is(getErr, ErrCacheMiss) {
			t.Fatalf("expected miss after expiry, got %v", getErr)
		}
	})

	t.Run("ristretto uses default ttl when ttl<=0", func(t *testing.T) {
		t.Parallel()

		// Exercise the ttl<=0 fallback path inside the ristretto setFn closure.
		c, err := NewCache[string, []byte](WithBackend(BackendRistretto), WithTTL(time.Minute))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = c.Stop(context.Background()) }()

		setErr := c.Set(context.Background(), "k", []byte("v"), 0)
		if setErr != nil {
			t.Fatalf("set failed: %v", setErr)
		}
	})

	t.Run("go-cache uses default ttl when ttl<=0", func(t *testing.T) {
		t.Parallel()

		c, err := NewCache[string, []byte](WithBackend(BackendGoCache), WithTTL(time.Minute))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = c.Stop(context.Background()) }()

		setErr := c.Set(context.Background(), "k", []byte("v"), 0)
		if setErr != nil {
			t.Fatalf("set failed: %v", setErr)
		}
	})
}

func TestCache_Stop(t *testing.T) {
	t.Parallel()

	t.Run("releases resources", func(t *testing.T) {
		t.Parallel()

		c, err := NewCache[string, []byte](WithBackend(BackendBigcache))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		stopErr := c.Stop(context.Background())
		if stopErr != nil {
			t.Fatalf("stop failed: %v", stopErr)
		}
	})

	t.Run("idempotent", func(t *testing.T) {
		t.Parallel()

		c, err := NewCache[string, []byte](WithBackend(BackendRistretto))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_ = c.Stop(context.Background())
		secondErr := c.Stop(context.Background())
		if secondErr != nil {
			t.Fatalf("second stop should be a no-op, got: %v", secondErr)
		}
	})
}

func TestCache_SlogEnabled(t *testing.T) {
	t.Parallel()

	c, err := NewCache[string, []byte](WithBackend(BackendGoCache), WithSlog())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()
	setErr := c.Set(ctx, "k", []byte("v"), time.Minute)
	if setErr != nil {
		t.Fatalf("set failed: %v", setErr)
	}

	_, getErr := c.Get(ctx, "k")
	if getErr != nil {
		t.Fatalf("get failed: %v", getErr)
	}

	_, missErr := c.Get(ctx, "missing")
	if !errors.Is(missErr, ErrCacheMiss) {
		t.Fatalf("expected miss, got %v", missErr)
	}

	delErr := c.Delete(ctx, "k")
	if delErr != nil {
		t.Fatalf("delete failed: %v", delErr)
	}

	clearErr := c.Clear(ctx)
	if clearErr != nil {
		t.Fatalf("clear failed: %v", clearErr)
	}

	// Stop with slog enabled to exercise recordStopped.
	stopErr := c.Stop(ctx)
	if stopErr != nil {
		t.Fatalf("stop failed: %v", stopErr)
	}
}

func TestCache_OTelMetrics(t *testing.T) {
	// Cannot be parallel: mutates the global MeterProvider.

	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))

	previous := otel.GetMeterProvider()
	otel.SetMeterProvider(provider)
	defer otel.SetMeterProvider(previous)

	c, err := NewCache[string, []byte](WithBackend(BackendGoCache), WithOTel(), WithOTelMeterName("cache-test"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = c.Stop(context.Background()) }()

	ctx := context.Background()
	setErr := c.Set(ctx, "k", []byte("v"), time.Minute)
	if setErr != nil {
		t.Fatalf("set failed: %v", setErr)
	}

	_, hitErr := c.Get(ctx, "k")
	if hitErr != nil {
		t.Fatalf("get failed: %v", hitErr)
	}

	_, missErr := c.Get(ctx, "missing")
	if !errors.Is(missErr, ErrCacheMiss) {
		t.Fatalf("expected miss, got %v", missErr)
	}

	delErr := c.Delete(ctx, "k")
	if delErr != nil {
		t.Fatalf("delete failed: %v", delErr)
	}

	rm := metricdata.ResourceMetrics{}
	collectErr := reader.Collect(context.Background(), &rm)
	if collectErr != nil {
		t.Fatalf("collect failed: %v", collectErr)
	}

	seen := map[string]bool{
		MetricHits:      false,
		MetricMisses:    false,
		MetricSets:      false,
		MetricEvictions: false,
	}

	for _, scope := range rm.ScopeMetrics {
		for _, m := range scope.Metrics {
			_, ok := seen[m.Name]
			if ok {
				seen[m.Name] = true
			}
		}
	}

	for name, present := range seen {
		if !present {
			t.Fatalf("metric %q not recorded", name)
		}
	}
}

func TestAssertValue(t *testing.T) {
	t.Parallel()

	t.Run("matching type returns value", func(t *testing.T) {
		t.Parallel()

		got, err := assertValue[[]byte](any([]byte("hi")))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != "hi" {
			t.Fatalf("got %q, want hi", got)
		}
	})

	t.Run("mismatched type returns ErrSerialization", func(t *testing.T) {
		t.Parallel()

		_, err := assertValue[[]byte](any(42))
		if !errors.Is(err, ErrSerialization) {
			t.Fatalf("expected ErrSerialization, got %v", err)
		}
	})
}

func TestCache_IntKey(t *testing.T) {
	t.Parallel()

	c, err := NewCache[int, []byte](WithBackend(BackendGoCache))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = c.Stop(context.Background()) }()

	ctx := context.Background()
	setErr := c.Set(ctx, 7, []byte("seven"), time.Minute)
	if setErr != nil {
		t.Fatalf("set failed: %v", setErr)
	}

	got, getErr := c.Get(ctx, 7)
	if getErr != nil {
		t.Fatalf("get failed: %v", getErr)
	}
	if string(got) != "seven" {
		t.Fatalf("got %q, want seven", got)
	}
}

