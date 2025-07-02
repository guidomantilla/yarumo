package boot

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/guidomantilla/yarumo/pkg/common/assert"
	clog "github.com/guidomantilla/yarumo/pkg/common/log"
	"github.com/guidomantilla/yarumo/pkg/common/pointer"
	"github.com/guidomantilla/yarumo/pkg/common/utils"
)

var singleton atomic.Value

type WireContext[C any] struct {
	Container
	Config C
}

func Context[C any]() (*WireContext[C], error) {
	value := singleton.Load()
	if utils.Empty(value) {
		return nil, errors.New("server - error getting context: context is nil")
	}

	if wctx, ok := value.(*WireContext[C]); ok {
		return wctx, nil
	}

	return nil, errors.New("server - error getting context: context is not of type WireContext")
}

func NewWireContext[C any](name string, version string, opts ...Option) *WireContext[C] {
	assert.NotEmpty(name, fmt.Sprintf("%s - error creating: appName is empty", "application"))
	assert.NotEmpty(version, fmt.Sprintf("%s - error creating: appName is empty", "application"))

	container := &Container{
		opts:       opts,
		AppName:    name,
		AppVersion: version,
		Config:     pointer.Zero[C](),
		Logger:     clog.Configure(name, version),
		Validator:  validator.New(),
	}

	viper.AutomaticEnv()
	options := NewOptions(container.opts...)

	log.Info().Str("stage", "startup").Str("component", "application").Msg("starting")
	defer log.Info().Str("stage", "startup").Str("component", "application").Msg("started")

	options.Logger(container)
	log.Info().Str("stage", "startup").Str("component", "application").Msg("logger set up")

	options.Config(container)
	log.Info().Str("stage", "startup").Str("component", "application").Msg("configuration set up")

	options.Validator(container)
	log.Info().Str("stage", "startup").Str("component", "application").Msg("validator set up")

	options.PasswordEncoder(container)
	log.Info().Str("stage", "startup").Str("component", "application").Msg("password encoder set up")

	options.PasswordGenerator(container)
	log.Info().Str("stage", "startup").Str("component", "application").Msg("password generator set up")

	options.TokenGenerator(container)
	log.Info().Str("stage", "startup").Str("component", "application").Msg("token generator set up")

	wctx := &WireContext[C]{
		Container: *container,
		Config:    container.Config.(C),
	}
	singleton.Store(wctx)
	return wctx
}

func (wctx *WireContext[T]) Stop(ctx context.Context) {
	assert.NotNil(ctx, fmt.Sprintf("%s -  error shutting down: context is nil", "application"))

	log.Info().Str("stage", "shut down").Str("component", "application").Msg("stopping")
	defer log.Info().Str("stage", "shut down").Str("component", "application").Msg("stopped")
}
