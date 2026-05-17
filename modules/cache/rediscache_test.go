package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"

	ccache "github.com/guidomantilla/yarumo/common/cache"
)

// jsonStringV is the JSON encoding of the string "v" used across Set/Get
// fixtures in this file.
const jsonStringV = `"v"`

func newTestRedisCache(t *testing.T) (ccache.Cache[string, string], *miniredis.Miniredis) {
	t.Helper()

	server := miniredis.RunT(t)

	c, err := BuildRedisCache[string]("test", WithRedisAddr(server.Addr()))
	if err != nil {
		t.Fatalf("BuildRedisCache: %v", err)
	}

	t.Cleanup(func() { _ = c.Stop(context.Background()) })

	return c, server
}

func TestBuildRedisCache(t *testing.T) {
	t.Parallel()

	t.Run("returns a usable cache when redis responds to ping", func(t *testing.T) {
		t.Parallel()

		c, _ := newTestRedisCache(t)

		ctx := context.Background()
		err := c.Set(ctx, "k", "v", 0)
		if err != nil {
			t.Fatalf("Set: %v", err)
		}

		got, getErr := c.Get(ctx, "k")
		if getErr != nil {
			t.Fatalf("Get: %v", getErr)
		}

		if got != "v" {
			t.Fatalf("Get = %q, want %q", got, "v")
		}
	})

	t.Run("fails fast when ping is unreachable and lazy init is off", func(t *testing.T) {
		t.Parallel()

		// 127.0.0.1:1 is a TCP port reserved as 'tcpmux' (rarely listening);
		// go-redis dial timeout will surface a connection error quickly.
		_, err := BuildRedisCache[string]("offline", WithRedisAddr("127.0.0.1:1"))
		if err == nil {
			t.Fatal("expected ping error for unreachable redis")
		}

		if !errors.Is(err, ErrRedisCommandFailed) {
			t.Fatalf("expected wrap of ErrRedisCommandFailed, got %v", err)
		}
	})

	t.Run("WithLazyInit skips the ping, so unreachable addr does not fail Build", func(t *testing.T) {
		t.Parallel()

		c, err := BuildRedisCache[string]("offline-lazy",
			WithRedisAddr("127.0.0.1:1"),
			WithLazyInit(),
		)
		if err != nil {
			t.Fatalf("BuildRedisCache with WithLazyInit: %v", err)
		}
		t.Cleanup(func() { _ = c.Stop(context.Background()) })

		// First command surfaces the connection error.
		_, getErr := c.Get(context.Background(), "k")
		if getErr == nil {
			t.Fatal("expected command error on first call against unreachable addr")
		}

		if !errors.Is(getErr, ErrRedisCommandFailed) {
			t.Fatalf("expected wrap of ErrRedisCommandFailed, got %v", getErr)
		}
	})
}

func TestRedisCache_Name(t *testing.T) {
	t.Parallel()

	t.Run("returns the configured name", func(t *testing.T) {
		t.Parallel()

		server := miniredis.RunT(t)

		c, err := BuildRedisCache[string]("alpha", WithRedisAddr(server.Addr()))
		if err != nil {
			t.Fatalf("BuildRedisCache: %v", err)
		}
		t.Cleanup(func() { _ = c.Stop(context.Background()) })

		if c.Name() != "alpha" {
			t.Fatalf("Name() = %q, want %q", c.Name(), "alpha")
		}
	})
}

func TestRedisCache_Get(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("returns stored value on hit", func(t *testing.T) {
		t.Parallel()

		c, _ := newTestRedisCache(t)

		setErr := c.Set(ctx, "k", "v", 0)
		if setErr != nil {
			t.Fatalf("Set: %v", setErr)
		}

		got, getErr := c.Get(ctx, "k")
		if getErr != nil {
			t.Fatalf("Get: %v", getErr)
		}

		if got != "v" {
			t.Fatalf("Get = %q, want %q", got, "v")
		}
	})

	t.Run("returns ErrCacheMiss when key absent", func(t *testing.T) {
		t.Parallel()

		c, _ := newTestRedisCache(t)

		got, err := c.Get(ctx, "missing")
		if err == nil {
			t.Fatal("expected error for missing key")
		}

		if !errors.Is(err, ccache.ErrCacheMiss) {
			t.Fatalf("expected wrap of ErrCacheMiss, got %v", err)
		}

		if got != "" {
			t.Fatalf("expected zero value on miss, got %q", got)
		}
	})

	t.Run("returns ErrRedisCommandFailed when server is unreachable", func(t *testing.T) {
		t.Parallel()

		c, server := newTestRedisCache(t)
		server.Close()

		_, err := c.Get(ctx, "k")
		if err == nil {
			t.Fatal("expected command error against closed server")
		}

		if !errors.Is(err, ErrRedisCommandFailed) {
			t.Fatalf("expected wrap of ErrRedisCommandFailed, got %v", err)
		}
	})

	t.Run("returns ErrRedisDecodeFailed when stored bytes do not decode into V", func(t *testing.T) {
		t.Parallel()

		server := miniredis.RunT(t)

		c, err := BuildRedisCache[int]("decode-test", WithRedisAddr(server.Addr()))
		if err != nil {
			t.Fatalf("BuildRedisCache: %v", err)
		}
		t.Cleanup(func() { _ = c.Stop(ctx) })

		server.Set("decode-test:k", "not-an-int")

		_, err = c.Get(ctx, "k")
		if err == nil {
			t.Fatal("expected decode error")
		}

		if !errors.Is(err, ErrRedisDecodeFailed) {
			t.Fatalf("expected wrap of ErrRedisDecodeFailed, got %v", err)
		}
	})
}

func TestRedisCache_Set(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("stores value with per-call ttl", func(t *testing.T) {
		t.Parallel()

		c, server := newTestRedisCache(t)

		err := c.Set(ctx, "k", "v", 10*time.Second)
		if err != nil {
			t.Fatalf("Set: %v", err)
		}

		raw, getErr := server.Get("test:k")
		if getErr != nil {
			t.Fatalf("server.Get: %v", getErr)
		}

		if raw != jsonStringV {
			t.Fatalf("raw = %q, want JSON-encoded %q", raw, jsonStringV)
		}
	})

	t.Run("falls back to default ttl when per-call ttl is non-positive", func(t *testing.T) {
		t.Parallel()

		server := miniredis.RunT(t)
		c, err := BuildRedisCache[string]("default-ttl-test",
			WithRedisAddr(server.Addr()),
			WithTTL(30*time.Second),
		)
		if err != nil {
			t.Fatalf("BuildRedisCache: %v", err)
		}
		t.Cleanup(func() { _ = c.Stop(ctx) })

		setErr := c.Set(ctx, "k", "v", 0)
		if setErr != nil {
			t.Fatalf("Set: %v", setErr)
		}

		ttl := server.TTL("default-ttl-test:k")
		if ttl != 30*time.Second {
			t.Fatalf("server TTL = %v, want %v", ttl, 30*time.Second)
		}
	})

	t.Run("returns ErrRedisEncodeFailed when codec cannot encode", func(t *testing.T) {
		t.Parallel()

		server := miniredis.RunT(t)
		c, err := BuildRedisCache[chan int]("encode-test", WithRedisAddr(server.Addr()))
		if err != nil {
			t.Fatalf("BuildRedisCache: %v", err)
		}
		t.Cleanup(func() { _ = c.Stop(ctx) })

		setErr := c.Set(ctx, "k", make(chan int), 0)
		if setErr == nil {
			t.Fatal("expected encode error")
		}

		if !errors.Is(setErr, ErrRedisEncodeFailed) {
			t.Fatalf("expected wrap of ErrRedisEncodeFailed, got %v", setErr)
		}
	})

	t.Run("returns ErrRedisCommandFailed when server is unreachable", func(t *testing.T) {
		t.Parallel()

		c, server := newTestRedisCache(t)
		server.Close()

		err := c.Set(ctx, "k", "v", 0)
		if err == nil {
			t.Fatal("expected command error against closed server")
		}

		if !errors.Is(err, ErrRedisCommandFailed) {
			t.Fatalf("expected wrap of ErrRedisCommandFailed, got %v", err)
		}
	})
}

func TestRedisCache_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("removes stored value", func(t *testing.T) {
		t.Parallel()

		c, _ := newTestRedisCache(t)

		setErr := c.Set(ctx, "k", "v", 0)
		if setErr != nil {
			t.Fatalf("Set: %v", setErr)
		}

		delErr := c.Delete(ctx, "k")
		if delErr != nil {
			t.Fatalf("Delete: %v", delErr)
		}

		has, hasErr := c.Has(ctx, "k")
		if hasErr != nil {
			t.Fatalf("Has: %v", hasErr)
		}

		if has {
			t.Fatal("expected key to be absent after Delete")
		}
	})

	t.Run("is a no-op when key absent", func(t *testing.T) {
		t.Parallel()

		c, _ := newTestRedisCache(t)

		err := c.Delete(ctx, "missing")
		if err != nil {
			t.Fatalf("Delete on missing key: %v", err)
		}
	})

	t.Run("returns ErrRedisCommandFailed when server is unreachable", func(t *testing.T) {
		t.Parallel()

		c, server := newTestRedisCache(t)
		server.Close()

		err := c.Delete(ctx, "k")
		if err == nil {
			t.Fatal("expected command error against closed server")
		}

		if !errors.Is(err, ErrRedisCommandFailed) {
			t.Fatalf("expected wrap of ErrRedisCommandFailed, got %v", err)
		}
	})
}

func TestRedisCache_Has(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("returns true for present key", func(t *testing.T) {
		t.Parallel()

		c, _ := newTestRedisCache(t)

		setErr := c.Set(ctx, "k", "v", 0)
		if setErr != nil {
			t.Fatalf("Set: %v", setErr)
		}

		has, hasErr := c.Has(ctx, "k")
		if hasErr != nil {
			t.Fatalf("Has: %v", hasErr)
		}

		if !has {
			t.Fatal("Has = false, want true")
		}
	})

	t.Run("returns false for absent key", func(t *testing.T) {
		t.Parallel()

		c, _ := newTestRedisCache(t)

		has, hasErr := c.Has(ctx, "missing")
		if hasErr != nil {
			t.Fatalf("Has: %v", hasErr)
		}

		if has {
			t.Fatal("Has = true, want false")
		}
	})

	t.Run("returns ErrRedisCommandFailed when server is unreachable", func(t *testing.T) {
		t.Parallel()

		c, server := newTestRedisCache(t)
		server.Close()

		_, err := c.Has(ctx, "k")
		if err == nil {
			t.Fatal("expected command error against closed server")
		}

		if !errors.Is(err, ErrRedisCommandFailed) {
			t.Fatalf("expected wrap of ErrRedisCommandFailed, got %v", err)
		}
	})
}

func TestRedisCache_Clear(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("removes only entries under this cache's key prefix", func(t *testing.T) {
		t.Parallel()

		server := miniredis.RunT(t)

		c, err := BuildRedisCache[string]("scoped", WithRedisAddr(server.Addr()))
		if err != nil {
			t.Fatalf("BuildRedisCache: %v", err)
		}
		t.Cleanup(func() { _ = c.Stop(ctx) })

		setErr := c.Set(ctx, "k1", "v1", 0)
		if setErr != nil {
			t.Fatalf("Set k1: %v", setErr)
		}

		setErr = c.Set(ctx, "k2", "v2", 0)
		if setErr != nil {
			t.Fatalf("Set k2: %v", setErr)
		}

		server.Set("other:keep", "x")

		clearErr := c.Clear(ctx)
		if clearErr != nil {
			t.Fatalf("Clear: %v", clearErr)
		}

		has1, _ := c.Has(ctx, "k1")
		has2, _ := c.Has(ctx, "k2")
		if has1 || has2 {
			t.Fatal("expected cache to be empty after Clear")
		}

		survivor, getErr := server.Get("other:keep")
		if getErr != nil {
			t.Fatalf("server.Get: %v", getErr)
		}

		if survivor != "x" {
			t.Fatalf("survivor = %q, want %q", survivor, "x")
		}
	})

	t.Run("returns ErrRedisCommandFailed when server is unreachable", func(t *testing.T) {
		t.Parallel()

		c, server := newTestRedisCache(t)
		server.Close()

		err := c.Clear(ctx)
		if err == nil {
			t.Fatal("expected command error against closed server")
		}

		if !errors.Is(err, ErrRedisCommandFailed) {
			t.Fatalf("expected wrap of ErrRedisCommandFailed, got %v", err)
		}
	})
}

func TestRedisCache_Stop(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("closes the client and is safe to call twice", func(t *testing.T) {
		t.Parallel()

		server := miniredis.RunT(t)
		c, err := BuildRedisCache[string]("stop-test", WithRedisAddr(server.Addr()))
		if err != nil {
			t.Fatalf("BuildRedisCache: %v", err)
		}

		stopErr := c.Stop(ctx)
		if stopErr != nil {
			t.Fatalf("first Stop: %v", stopErr)
		}

		stopErr = c.Stop(ctx)
		if stopErr != nil {
			t.Fatalf("second Stop: %v", stopErr)
		}
	})
}

func TestRedisCache_KeyPrefix(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("default prefix uses name", func(t *testing.T) {
		t.Parallel()

		server := miniredis.RunT(t)
		c, err := BuildRedisCache[string]("alpha", WithRedisAddr(server.Addr()))
		if err != nil {
			t.Fatalf("BuildRedisCache: %v", err)
		}
		t.Cleanup(func() { _ = c.Stop(ctx) })

		setErr := c.Set(ctx, "k", "v", 0)
		if setErr != nil {
			t.Fatalf("Set: %v", setErr)
		}

		raw, getErr := server.Get("alpha:k")
		if getErr != nil {
			t.Fatalf("server.Get(alpha:k): %v", getErr)
		}

		if raw != jsonStringV {
			t.Fatalf("raw = %q, want JSON-encoded %q", raw, jsonStringV)
		}
	})

	t.Run("WithKeyPrefix overrides default", func(t *testing.T) {
		t.Parallel()

		server := miniredis.RunT(t)
		c, err := BuildRedisCache[string]("alpha",
			WithRedisAddr(server.Addr()),
			WithKeyPrefix("custom::"),
		)
		if err != nil {
			t.Fatalf("BuildRedisCache: %v", err)
		}
		t.Cleanup(func() { _ = c.Stop(ctx) })

		setErr := c.Set(ctx, "k", "v", 0)
		if setErr != nil {
			t.Fatalf("Set: %v", setErr)
		}

		raw, getErr := server.Get("custom::k")
		if getErr != nil {
			t.Fatalf("server.Get(custom::k): %v", getErr)
		}

		if raw != jsonStringV {
			t.Fatalf("raw = %q, want JSON-encoded %q", raw, jsonStringV)
		}
	})
}
