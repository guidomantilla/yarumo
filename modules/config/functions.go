package config

import (
	"context"
	"log/slog"
	"os"
	"strings"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	clog "github.com/guidomantilla/yarumo/common/log"
	cslog "github.com/guidomantilla/yarumo/common/log/slog"
	cutils "github.com/guidomantilla/yarumo/common/utils"
	"github.com/spf13/viper"
)

// Default configures the application's cross-cutting concerns: environment variable loading,
// assertion subsystem, and logging. ctx must be non-nil; it is returned unchanged.
func Default(ctx context.Context, name string, version string, env string) context.Context {
	cassert.NotNil(ctx, "ctx is nil")

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
