package redis

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

		if opts.addr != "" {
			t.Fatalf("addr = %q, want empty", opts.addr)
		}

		if opts.password != "" {
			t.Fatalf("password = %q, want empty", opts.password)
		}

		if opts.db != 0 {
			t.Fatalf("db = %d, want 0", opts.db)
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
			WithAddr("redis.local:6379"),
		)

		if opts.ttl != 10*time.Minute {
			t.Fatalf("ttl = %v, want %v", opts.ttl, 10*time.Minute)
		}

		if opts.addr != "redis.local:6379" {
			t.Fatalf("addr = %q, want %q", opts.addr, "redis.local:6379")
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

func TestWithAddr(t *testing.T) {
	t.Parallel()

	t.Run("sets addr when non-empty", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithAddr("redis.local:6379"))
		if opts.addr != "redis.local:6379" {
			t.Fatalf("addr = %q, want %q", opts.addr, "redis.local:6379")
		}
	})

	t.Run("ignores empty addr, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithAddr(""))
		if opts.addr != "" {
			t.Fatalf("addr = %q, want empty", opts.addr)
		}
	})
}

func TestWithPassword(t *testing.T) {
	t.Parallel()

	t.Run("sets password when non-empty", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPassword("secret"))
		if opts.password != "secret" {
			t.Fatalf("password = %q, want %q", opts.password, "secret")
		}
	})

	t.Run("ignores empty password, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithPassword(""))
		if opts.password != "" {
			t.Fatalf("password = %q, want empty", opts.password)
		}
	})
}

func TestWithDB(t *testing.T) {
	t.Parallel()

	t.Run("sets DB index when non-negative", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDB(3))
		if opts.db != 3 {
			t.Fatalf("db = %d, want 3", opts.db)
		}
	})

	t.Run("accepts zero DB index", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDB(0))
		if opts.db != 0 {
			t.Fatalf("db = %d, want 0", opts.db)
		}
	})

	t.Run("ignores negative DB index, preserves default", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithDB(-1))
		if opts.db != 0 {
			t.Fatalf("db = %d, want default 0", opts.db)
		}
	})
}

type stubCodec struct{}

func (stubCodec) Encode(any) ([]byte, error) { return nil, nil }
func (stubCodec) Decode([]byte, any) error   { return nil }

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
