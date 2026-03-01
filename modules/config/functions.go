// Package config provides application bootstrap configuration.
package config

import (
	"context"
	"log/slog"
	"os"
	"sort"
	"strings"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	clog "github.com/guidomantilla/yarumo/common/log"
	cslog "github.com/guidomantilla/yarumo/common/log/slog"
	cutils "github.com/guidomantilla/yarumo/common/utils"
	"github.com/spf13/viper"
)

// Default configures the application's cross-cutting concerns: environment variable loading,
// assertion subsystem, and logging.
func Default(ctx context.Context, name string, version string, env string) context.Context {

	viper.AutomaticEnv()

	v := strings.ToLower(viper.GetString("ENABLE_ASSERTS"))
	cassert.Enable(v == "1" || v == "true" || v == "yes")

	level := parseLevel(cutils.Coalesce(viper.GetString("LOG_LEVEL"), "info"))

	handlerOpts := &slog.HandlerOptions{
		AddSource:   viper.GetBool("DEBUG"),
		Level:       slog.Level(level),
		ReplaceAttr: cslog.ReplaceLevel,
	}

	var handler slog.Handler = slog.NewJSONHandler(os.Stderr, handlerOpts)

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
		handler = handler.WithAttrs(attrs)
	}

	clog.Use(cslog.NewLogger(cslog.WithHandlers(handler)))

	v = strings.ToLower(viper.GetString("ENABLE_CONFIG_DUMP"))
	if v == "1" || v == "true" || v == "yes" {
		dump(ctx)
	}

	return ctx
}

const maskedValue = "********"

func dump(ctx context.Context) {

	envs := os.Environ()
	sort.Strings(envs)

	args := make([]any, 0, len(envs)*2)
	for _, env := range envs {
		parts := strings.SplitN(env, "=", 2)
		key := parts[0]
		value := parts[1]

		if shouldMask(key) {
			value = maskValue(value)
		}

		args = append(args, key, value)
	}

	clog.Info(ctx, "config dump", args...)
}

func shouldMask(key string) bool {
	upper := strings.ToUpper(key)

	return strings.Contains(upper, "PASSWORD") ||
		strings.Contains(upper, "SECRET") ||
		strings.Contains(upper, "TOKEN") ||
		strings.Contains(upper, "KEY") ||
		strings.Contains(upper, "CREDENTIAL") ||
		strings.Contains(upper, "PRIVATE")
}

func maskValue(_ string) string {
	return maskedValue
}

func parseLevel(s string) cslog.Level {
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
