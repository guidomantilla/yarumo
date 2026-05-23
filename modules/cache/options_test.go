package cache

import (
	"testing"
	"time"

	ccache "github.com/guidomantilla/yarumo/common/cache"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("applies safe defaults when no options given", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()

		if opts.ttl != 5*time.Minute {
			t.Fatalf("ttl = %v, want %v", opts.ttl, 5*time.Minute)
		}

		if opts.keyPrefix != "" {
			t.Fatalf("keyPrefix = %q, want empty", opts.keyPrefix)
		}

		if opts.ristrettoNumCtrs != 1_000_000 {
			t.Fatalf("ristrettoNumCtrs = %d, want %d", opts.ristrettoNumCtrs, 1_000_000)
		}

		if opts.ristrettoMaxCost != 100<<20 {
			t.Fatalf("ristrettoMaxCost = %d, want %d", opts.ristrettoMaxCost, 100<<20)
		}

		if opts.ristrettoBufItems != 64 {
			t.Fatalf("ristrettoBufItems = %d, want %d", opts.ristrettoBufItems, 64)
		}

		if opts.redisAddr != "" {
			t.Fatalf("redisAddr = %q, want empty", opts.redisAddr)
		}

		if opts.redisPassword != "" {
			t.Fatalf("redisPassword = %q, want empty", opts.redisPassword)
		}

		if opts.redisDB != 0 {
			t.Fatalf("redisDB = %d, want 0", opts.redisDB)
		}

		_, ok := opts.codec.(ccache.JSONCodec)
		if !ok {
			t.Fatalf("codec = %T, want JSONCodec", opts.codec)
		}
	})

	t.Run("applies each option in order", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(
			WithTTL(10*time.Minute),
			WithRistrettoCapacity(2_000_000, 200<<20, 128),
		)

		if opts.ttl != 10*time.Minute {
			t.Fatalf("ttl = %v, want %v", opts.ttl, 10*time.Minute)
		}

		if opts.ristrettoNumCtrs != 2_000_000 {
			t.Fatalf("ristrettoNumCtrs = %d, want %d", opts.ristrettoNumCtrs, 2_000_000)
		}
	})
}

func TestWithTTL(t *testing.T) {
	t.Parallel()

	t.Run("sets the ttl when positive", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTTL(30 * time.Second))
		if opts.ttl != 30*time.Second {
			t.Fatalf("ttl = %v, want %v", opts.ttl, 30*time.Second)
		}
	})

	t.Run("ignores zero ttl, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTTL(0))
		if opts.ttl != 5*time.Minute {
			t.Fatalf("ttl = %v, want default %v", opts.ttl, 5*time.Minute)
		}
	})

	t.Run("ignores negative ttl, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithTTL(-1 * time.Second))
		if opts.ttl != 5*time.Minute {
			t.Fatalf("ttl = %v, want default %v", opts.ttl, 5*time.Minute)
		}
	})
}

func TestWithRistrettoCapacity(t *testing.T) {
	t.Parallel()

	t.Run("sets every parameter when all positive", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRistrettoCapacity(500_000, 50<<20, 32))

		if opts.ristrettoNumCtrs != 500_000 {
			t.Fatalf("numCounters = %d, want %d", opts.ristrettoNumCtrs, 500_000)
		}

		if opts.ristrettoMaxCost != 50<<20 {
			t.Fatalf("maxCost = %d, want %d", opts.ristrettoMaxCost, 50<<20)
		}

		if opts.ristrettoBufItems != 32 {
			t.Fatalf("bufferItems = %d, want %d", opts.ristrettoBufItems, 32)
		}
	})

	t.Run("ignores non-positive numCounters", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRistrettoCapacity(0, 50<<20, 32))
		if opts.ristrettoNumCtrs != 1_000_000 {
			t.Fatalf("numCounters = %d, want default %d", opts.ristrettoNumCtrs, 1_000_000)
		}
	})

	t.Run("ignores non-positive maxCost", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRistrettoCapacity(500_000, 0, 32))
		if opts.ristrettoMaxCost != 100<<20 {
			t.Fatalf("maxCost = %d, want default %d", opts.ristrettoMaxCost, 100<<20)
		}
	})

	t.Run("ignores non-positive bufferItems", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRistrettoCapacity(500_000, 50<<20, 0))
		if opts.ristrettoBufItems != 64 {
			t.Fatalf("bufferItems = %d, want default %d", opts.ristrettoBufItems, 64)
		}
	})
}

func TestWithKeyPrefix(t *testing.T) {
	t.Parallel()

	t.Run("sets prefix when non-empty", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithKeyPrefix("svc::"))
		if opts.keyPrefix != "svc::" {
			t.Fatalf("keyPrefix = %q, want %q", opts.keyPrefix, "svc::")
		}
	})

	t.Run("ignores empty prefix, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithKeyPrefix(""))
		if opts.keyPrefix != "" {
			t.Fatalf("keyPrefix = %q, want empty", opts.keyPrefix)
		}
	})
}

func TestWithRedisAddr(t *testing.T) {
	t.Parallel()

	t.Run("sets addr when non-empty", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRedisAddr("redis.local:6379"))
		if opts.redisAddr != "redis.local:6379" {
			t.Fatalf("redisAddr = %q, want %q", opts.redisAddr, "redis.local:6379")
		}
	})

	t.Run("ignores empty addr, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRedisAddr(""))
		if opts.redisAddr != "" {
			t.Fatalf("redisAddr = %q, want empty", opts.redisAddr)
		}
	})
}

func TestWithRedisPassword(t *testing.T) {
	t.Parallel()

	t.Run("sets password when non-empty", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRedisPassword("secret"))
		if opts.redisPassword != "secret" {
			t.Fatalf("redisPassword = %q, want %q", opts.redisPassword, "secret")
		}
	})

	t.Run("ignores empty password, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRedisPassword(""))
		if opts.redisPassword != "" {
			t.Fatalf("redisPassword = %q, want empty", opts.redisPassword)
		}
	})
}

func TestWithRedisDB(t *testing.T) {
	t.Parallel()

	t.Run("sets DB index when non-negative", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRedisDB(3))
		if opts.redisDB != 3 {
			t.Fatalf("redisDB = %d, want 3", opts.redisDB)
		}
	})

	t.Run("accepts zero DB index", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRedisDB(0))
		if opts.redisDB != 0 {
			t.Fatalf("redisDB = %d, want 0", opts.redisDB)
		}
	})

	t.Run("ignores negative DB index, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithRedisDB(-1))
		if opts.redisDB != 0 {
			t.Fatalf("redisDB = %d, want default 0", opts.redisDB)
		}
	})
}

type stubCodec struct{}

func (stubCodec) Encode(any) ([]byte, error)   { return nil, nil }
func (stubCodec) Decode([]byte, any) error     { return nil }

func TestWithCodec(t *testing.T) {
	t.Parallel()

	t.Run("sets codec when non-nil", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithCodec(stubCodec{}))
		_, ok := opts.codec.(stubCodec)
		if !ok {
			t.Fatalf("codec = %T, want stubCodec", opts.codec)
		}
	})

	t.Run("ignores nil codec, preserves JSONCodec default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithCodec(nil))
		_, ok := opts.codec.(ccache.JSONCodec)
		if !ok {
			t.Fatalf("codec = %T, want JSONCodec", opts.codec)
		}
	})
}
