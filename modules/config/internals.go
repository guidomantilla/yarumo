package config

import (
	"context"
	"os"
	"sort"
	"strings"

	clog "github.com/guidomantilla/yarumo/log"
	cslog "github.com/guidomantilla/yarumo/log/slog"
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
