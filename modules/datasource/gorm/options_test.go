package gorm

import (
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	t.Run("uses safe defaults", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions()

		if opts.gormConfig == nil {
			t.Fatal("expected non-nil gormConfig default")
		}

		if opts.maxIdleConns != 10 {
			t.Fatalf("maxIdleConns = %d, want 10", opts.maxIdleConns)
		}

		if opts.maxOpenConns != 100 {
			t.Fatalf("maxOpenConns = %d, want 100", opts.maxOpenConns)
		}

		if opts.connMaxLifetime != 30*time.Minute {
			t.Fatalf("connMaxLifetime = %v, want 30m", opts.connMaxLifetime)
		}

		if opts.connMaxIdleTime != 5*time.Minute {
			t.Fatalf("connMaxIdleTime = %v, want 5m", opts.connMaxIdleTime)
		}
	})
}

func TestWithGormConfig(t *testing.T) {
	t.Parallel()

	t.Run("overrides the default config", func(t *testing.T) {
		t.Parallel()

		cfg := &gorm.Config{}

		opts := NewOptions(WithGormConfig(cfg))

		if opts.gormConfig != cfg {
			t.Fatal("expected WithGormConfig to replace default")
		}
	})

	t.Run("nil values are ignored", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithGormConfig(nil))

		if opts.gormConfig == nil {
			t.Fatal("expected nil override to be ignored")
		}
	})
}

func TestWithMaxIdleConns(t *testing.T) {
	t.Parallel()

	t.Run("applies positive values", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxIdleConns(7))
		if opts.maxIdleConns != 7 {
			t.Fatalf("maxIdleConns = %d, want 7", opts.maxIdleConns)
		}
	})

	t.Run("ignores non-positive values", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxIdleConns(0))
		if opts.maxIdleConns != 10 {
			t.Fatalf("expected default 10, got %d", opts.maxIdleConns)
		}
	})
}

func TestWithMaxOpenConns(t *testing.T) {
	t.Parallel()

	t.Run("applies positive values", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxOpenConns(50))
		if opts.maxOpenConns != 50 {
			t.Fatalf("maxOpenConns = %d, want 50", opts.maxOpenConns)
		}
	})

	t.Run("ignores non-positive values", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithMaxOpenConns(-1))
		if opts.maxOpenConns != 100 {
			t.Fatalf("expected default 100, got %d", opts.maxOpenConns)
		}
	})
}

func TestWithConnMaxLifetime(t *testing.T) {
	t.Parallel()

	t.Run("applies positive durations", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithConnMaxLifetime(time.Hour))
		if opts.connMaxLifetime != time.Hour {
			t.Fatalf("connMaxLifetime = %v, want 1h", opts.connMaxLifetime)
		}
	})

	t.Run("ignores non-positive durations", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithConnMaxLifetime(0))
		if opts.connMaxLifetime != 30*time.Minute {
			t.Fatalf("expected default 30m, got %v", opts.connMaxLifetime)
		}
	})
}

func TestWithConnMaxIdleTime(t *testing.T) {
	t.Parallel()

	t.Run("applies positive durations", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithConnMaxIdleTime(2 * time.Minute))
		if opts.connMaxIdleTime != 2*time.Minute {
			t.Fatalf("connMaxIdleTime = %v, want 2m", opts.connMaxIdleTime)
		}
	})

	t.Run("ignores non-positive durations", func(t *testing.T) {
		t.Parallel()

		opts := NewOptions(WithConnMaxIdleTime(-time.Second))
		if opts.connMaxIdleTime != 5*time.Minute {
			t.Fatalf("expected default 5m, got %v", opts.connMaxIdleTime)
		}
	})
}

func TestPostgresOpener(t *testing.T) {
	t.Parallel()

	t.Run("returns a non-nil dialector factory", func(t *testing.T) {
		t.Parallel()

		fn := PostgresOpener()
		if fn == nil {
			t.Fatal("expected non-nil OpenFn")
		}

		d := fn("host=localhost user=u password=p dbname=db port=5432 sslmode=disable")
		if d == nil {
			t.Fatal("expected non-nil dialector")
		}
	})
}

func TestSqliteOpener(t *testing.T) {
	t.Parallel()

	t.Run("returns a non-nil dialector factory", func(t *testing.T) {
		t.Parallel()

		fn := SqliteOpener()
		if fn == nil {
			t.Fatal("expected non-nil OpenFn")
		}

		d := fn(":memory:")
		if d == nil {
			t.Fatal("expected non-nil dialector")
		}
	})
}
