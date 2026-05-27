package config

import (
	"context"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/spf13/viper"

	clog "github.com/guidomantilla/yarumo/core/common/log"
	cutils "github.com/guidomantilla/yarumo/core/common/utils"
	cslog "github.com/guidomantilla/yarumo/extension/common/log/slog"
)

const maskedValue = "********"

// dump logs every environment variable as a key/value pair, masking values
// for keys that look like secrets.
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

// shouldMask reports whether key looks like it carries a secret and should be
// masked before logging.
func shouldMask(key string) bool {
	upper := strings.ToUpper(key)

	return strings.Contains(upper, "PASSWORD") ||
		strings.Contains(upper, "SECRET") ||
		strings.Contains(upper, "TOKEN") ||
		strings.Contains(upper, "KEY") ||
		strings.Contains(upper, "CREDENTIAL") ||
		strings.Contains(upper, "PRIVATE")
}

// maskValue returns the standard mask placeholder regardless of the input.
func maskValue(_ string) string {
	return maskedValue
}

// parseLevel maps a textual log level name to the matching cslog.Level.
// Unknown values silently fall back to cslog.LevelInfo.
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

// buildDefaultLogger constructs the slog-backed clog.Logger used when the
// caller does not pass WithLogger. It reads LOG_LEVEL and DEBUG from viper,
// emits JSON to os.Stderr, and attaches name/version/env as base attrs
// when non-empty.
func buildDefaultLogger(name string, version string, env string) clog.Logger {
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

	return cslog.NewLogger(
		cslog.WithHandlers(handler),
		cslog.WithContextExtractors(cslog.SlogctxExtractor),
	)
}
