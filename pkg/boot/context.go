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
	LogLevel   zerolog.Level
	Config     any
}

func NewWireContext(appName string, version string, opts ...Option) *WireContext {
	assert.NotEmpty(appName, fmt.Sprintf("%s - error creating: appName is empty", "application"))
	assert.NotEmpty(version, fmt.Sprintf("%s - error creating: appName is empty", "application"))

	log.Info().Str("stage", "startup").Str("component", "application").Msg("starting up")

	wctx := &WireContext{
		AppName:    appName,
		AppVersion: version,
		DebugMode:  false,
		LogLevel:   zerolog.InfoLevel,
	}

	options := NewOptions(opts...)

	log.Info().Str("stage", "startup").Str("component", "application").Msg("setting up configuration")
	wctx.Config = options.Config(wctx)

	return wctx
}

func (wctx *WireContext) Stop(ctx context.Context) {
	assert.NotNil(ctx, fmt.Sprintf("%s -  error shutting down: context is nil", "application"))

	log.Info().Str("stage", "shut down").Str("component", "application").Msg("stopping")

	log.Info().Str("stage", "shut down").Str("component", "application").Msg("stopped")
}
