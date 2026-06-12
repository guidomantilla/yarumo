package config

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	clog "github.com/guidomantilla/yarumo/core/common/log"
	cutils "github.com/guidomantilla/yarumo/core/common/utils"
	cslog "github.com/guidomantilla/yarumo/extension/common/log/slog"
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

func SlogLogger(name string, version string, env string, handlers ...slog.Handler) clog.Logger {
	level := ParseLevel(cutils.Coalesce(viper.GetString("LOG_LEVEL"), "info"))

	handlerOpts := &slog.HandlerOptions{
		AddSource:   viper.GetBool("DEBUG"),
		Level:       slog.Level(level),
		ReplaceAttr: cslog.ReplaceLevel,
	}

	stderrHandler := slog.NewJSONHandler(os.Stderr, handlerOpts)
	handlers = append(handlers, stderrHandler)

	var attrs []slog.Attr
	if cutils.NotEmpty(name) {
		attrs = append(attrs, slog.String("name", name))
	}
	if cutils.NotEmpty(version) {
		attrs = append(attrs, slog.String("version", version))
	}
	if cutils.NotEmpty(env) {
		attrs = append(attrs, slog.String("env", env))
	}

	if len(attrs) > 0 {
		for i, h := range handlers {
			handlers[i] = h.WithAttrs(attrs)
		}
	}

	return cslog.NewLogger(cslog.WithHandlers(handlers...), cslog.WithContextExtractors(cslog.SlogctxExtractor))
}

// ParseLevel maps a textual log level name to the matching cslog.Level.
// Unknown values silently fall back to cslog.LevelInfo.
func ParseLevel(s string) cslog.Level {
	switch strings.ToLower(s) {
	case "trace":
		return cslog.LevelTrace
	case "debug":
		return cslog.LevelDebug
	case "info":
		return cslog.LevelInfo
	case "warn", "warning":
		return cslog.LevelWarn
	case "error":
		return cslog.LevelError
	case "fatal":
		return cslog.LevelFatal
	case "off", "disabled":
		return cslog.LevelOff
	default:
		return cslog.LevelInfo
	}
}
