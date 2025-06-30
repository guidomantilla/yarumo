package boot

import (
	"context"
	"fmt"

	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/guidomantilla/yarumo/pkg/common/assert"
	clog "github.com/guidomantilla/yarumo/pkg/common/log"
	"github.com/guidomantilla/yarumo/pkg/common/pointer"
)

type WireContext struct {
	opts       []Option
	AppName    string
	AppVersion string
	DebugMode  bool
	LogLevel   zerolog.Level
	Config     any
	Logger     zerolog.Logger
	Validator  *validator.Validate
}

func NewWireContext[T any](name string, version string, opts ...Option) *WireContext {
	assert.NotEmpty(name, fmt.Sprintf("%s - error creating: appName is empty", "application"))
	assert.NotEmpty(version, fmt.Sprintf("%s - error creating: appName is empty", "application"))
	return &WireContext{
		opts:       opts,
		AppName:    name,
		AppVersion: version,
		DebugMode:  false,
		LogLevel:   zerolog.InfoLevel,
		Config:     pointer.Zero[T](),
		Logger:     clog.Configure(name, version),
		Validator:  validator.New(),
	}
}

func (wctx *WireContext) Start(ctx context.Context) {
	assert.NotNil(ctx, fmt.Sprintf("%s -  error starting up: context is nil", "application"))

	log.Info().Str("stage", "startup").Str("component", "application").Msg("starting")
	defer log.Info().Str("stage", "startup").Str("component", "application").Msg("started")

	options := NewOptions(wctx.opts...)
	options.Logger(wctx)
	log.Info().Str("stage", "startup").Str("component", "application").Msg("logger set up")

	options.Config(wctx)
	log.Info().Str("stage", "startup").Str("component", "application").Msg("configuration set up")

	options.Validator(wctx)
	log.Info().Str("stage", "startup").Str("component", "application").Msg("validator set up")

}

func (wctx *WireContext) Stop(ctx context.Context) {
	assert.NotNil(ctx, fmt.Sprintf("%s -  error shutting down: context is nil", "application"))

	log.Info().Str("stage", "shut down").Str("component", "application").Msg("stopping")
	defer log.Info().Str("stage", "shut down").Str("component", "application").Msg("stopped")
}
