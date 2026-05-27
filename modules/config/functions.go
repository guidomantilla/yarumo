package config

import (
	"context"
	"strings"

	"github.com/spf13/viper"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	clog "github.com/guidomantilla/yarumo/core/common/log"
)

// Default configures the application's cross-cutting concerns: environment
// variable loading, the assertion subsystem and logging. ctx must be
// non-nil; it is returned unchanged.
//
// By default the installed logger is the slog-backed clog.Logger built
// from LOG_LEVEL / DEBUG. Override it with WithLogger to inject any other
// clog.Logger implementation (useful for tests, alternative backends, or
// pre-configured loggers wired earlier in the bootstrap).
func Default(ctx context.Context, name string, version string, env string, opts ...Option) context.Context {
	cassert.NotNil(ctx, "ctx is nil")

	viper.AutomaticEnv()

	v := strings.ToLower(viper.GetString("ENABLE_ASSERTS"))
	cassert.Enable(v == "1" || v == "true" || v == "yes")

	options := NewOptions(name, version, env, opts...)
	
	clog.Use(options.logger)

	v = strings.ToLower(viper.GetString("ENABLE_CONFIG_DUMP"))
	if v == "1" || v == "true" || v == "yes" {
		dump(ctx)
	}

	return ctx
}
