// Package config provides application bootstrap configuration.
//
// Environment variables read by Default:
//
//	ENABLE_ASSERTS      Toggles cassert.Enable. Truthy: "1", "true", "yes"
//	                    (case-insensitive). Anything else (including unset)
//	                    leaves assertions disabled.
//	LOG_LEVEL           One of: trace, debug, info, warn, warning, error,
//	                    fatal, off, disabled (case-insensitive). Unknown
//	                    values silently fall back to "info". Empty/unset
//	                    defaults to "info".
//	DEBUG               Boolean parsed by viper. When true, the slog handler
//	                    sets AddSource so logs include source file/line.
//	ENABLE_CONFIG_DUMP  Toggles a full environment dump at startup (sensitive
//	                    keys masked). Same truthy values as ENABLE_ASSERTS.
package config

import (
	"context"
)

var (
	_ DefaultFn = Default
)

// DefaultFn is the function type for Default.
type DefaultFn func(ctx context.Context, name string, version string, env string) context.Context
