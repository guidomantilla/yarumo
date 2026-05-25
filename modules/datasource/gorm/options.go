package gorm

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/glebarez/sqlite"
)

// Option is a functional option for configuring Options.
type Option func(opts *Options)

// Options holds the configuration applied at connection construction
// time.
type Options struct {
	gormConfig      *gorm.Config
	maxIdleConns    int
	maxOpenConns    int
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
}

// NewOptions constructs Options with safe defaults and applies the
// given functional options. Defaults:
//   - silent GORM logger to keep tests quiet; consumers swap in via
//     WithGormConfig.
//   - 10 idle / 100 open connections, 30-minute conn lifetime,
//     5-minute idle timeout (sql/database defaults are 2/0/0/0 which is
//     overly conservative for service workloads).
func NewOptions(opts ...Option) *Options {
	options := &Options{
		gormConfig:      &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)},
		maxIdleConns:    10,
		maxOpenConns:    100,
		connMaxLifetime: 30 * time.Minute,
		connMaxIdleTime: 5 * time.Minute,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithGormConfig overrides the *gorm.Config used to call gorm.Open. Nil
// values are ignored, preserving the default.
func WithGormConfig(cfg *gorm.Config) Option {
	return func(opts *Options) {
		if cfg != nil {
			opts.gormConfig = cfg
		}
	}
}

// WithMaxIdleConns sets the maximum number of idle pool connections.
// Non-positive values are ignored, preserving the default.
func WithMaxIdleConns(n int) Option {
	return func(opts *Options) {
		if n > 0 {
			opts.maxIdleConns = n
		}
	}
}

// WithMaxOpenConns sets the maximum number of open pool connections.
// Non-positive values are ignored, preserving the default.
func WithMaxOpenConns(n int) Option {
	return func(opts *Options) {
		if n > 0 {
			opts.maxOpenConns = n
		}
	}
}

// WithConnMaxLifetime sets the maximum amount of time a connection may
// be reused. Non-positive values are ignored, preserving the default.
func WithConnMaxLifetime(d time.Duration) Option {
	return func(opts *Options) {
		if d > 0 {
			opts.connMaxLifetime = d
		}
	}
}

// WithConnMaxIdleTime sets the maximum amount of time a connection may
// remain idle before being closed by the pool. Non-positive values are
// ignored, preserving the default.
func WithConnMaxIdleTime(d time.Duration) Option {
	return func(opts *Options) {
		if d > 0 {
			opts.connMaxIdleTime = d
		}
	}
}

// PostgresOpener returns an OpenFn that produces a postgres dialector
// from the given DSN. Convenience helper for the common case.
func PostgresOpener() OpenFn {
	return postgres.Open
}

// SqliteOpener returns an OpenFn that produces a sqlite dialector
// (pure-Go driver) from the given DSN. Convenience helper used by
// tests and lightweight deployments.
func SqliteOpener() OpenFn {
	return sqlite.Open
}
