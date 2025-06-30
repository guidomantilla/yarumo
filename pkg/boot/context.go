package boot

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/guidomantilla/yarumo/pkg/common/assert"
)

type WireContext struct {
	AppName    string
	AppVersion string
	DebugMode  bool
	Config     any
	LogLevel   zerolog.Level
}

func NewWireContext(name string, version string, opts ...Option) *WireContext {
	assert.NotEmpty(name, fmt.Sprintf("%s - error creating: appName is empty", "application"))
	assert.NotEmpty(version, fmt.Sprintf("%s - error creating: appName is empty", "application"))

	wctx := &WireContext{
		AppName:    name,
		AppVersion: version,
		DebugMode:  false,
		LogLevel:   zerolog.InfoLevel,
	}

	options := NewOptions(opts...)
	wctx.Config = options.Config(wctx)
	log.Info().Str("stage", "startup").Str("component", "application").Msg("starting up")
	log.Info().Str("stage", "startup").Str("component", "application").Msg("setting up configuration")

	return wctx
}

func (wctx *WireContext) Stop(ctx context.Context) {
	assert.NotNil(ctx, fmt.Sprintf("%s -  error shutting down: context is nil", "application"))

	log.Info().Str("stage", "shut down").Str("component", "application").Msg("stopping")

	log.Info().Str("stage", "shut down").Str("component", "application").Msg("stopped")
}
