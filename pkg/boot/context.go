package boot

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/guidomantilla/yarumo/pkg/common/assert"
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

func NewWireContext[C any](name string, version string, opts ...WireContextOption) *WireContext[C] {
	assert.NotEmpty(name, fmt.Sprintf("%s - error creating: appName is empty", "context"))
	assert.NotEmpty(version, fmt.Sprintf("%s - error creating: appName is empty", "context"))

	viper.AutomaticEnv()
	options := NewOptions(opts...)
	container := NewContainer[C]()

	log.Info().Str("stage", "startup").Str("component", "context").Msg("starting")
	defer log.Info().Str("stage", "startup").Str("component", "context").Msg("started")

	options.Hasher(container)
	log.Info().Str("stage", "startup").Str("component", "context").Msg("hasher set up")

	options.UIDGen(container)
	log.Info().Str("stage", "startup").Str("component", "context").Msg("uid generator set up")

	options.Logger(container)
	log.Info().Str("stage", "startup").Str("component", "context").Msg("logger set up")

	options.Config(container)
	log.Info().Str("stage", "startup").Str("component", "context").Msg("configuration set up")

	options.Validator(container)
	log.Info().Str("stage", "startup").Str("component", "context").Msg("validator set up")

	options.PasswordEncoder(container)
	log.Info().Str("stage", "startup").Str("component", "context").Msg("password encoder set up")

	options.PasswordGenerator(container)
	log.Info().Str("stage", "startup").Str("component", "context").Msg("password generator set up")

	options.TokenGenerator(container)
	log.Info().Str("stage", "startup").Str("component", "context").Msg("token generator set up")

	options.Cipher(container)
	log.Info().Str("stage", "startup").Str("component", "context").Msg("cipher set up")

	options.HttpClient(container)
	log.Info().Str("stage", "startup").Str("component", "context").Msg("http client set up")

	for _, beanFn := range options.More {
		if !utils.Empty(beanFn) {
			beanFn(container)
		}
	}

	wctx := &WireContext[C]{
		Container: *container,
		Config:    container.Config.(C),
	}
	singleton.Store(wctx)
	return wctx
}

func (wctx *WireContext[T]) Stop(ctx context.Context) {
	assert.NotNil(ctx, fmt.Sprintf("%s -  error shutting down: context is nil", "context"))

	log.Info().Str("stage", "shut down").Str("component", "context").Msg("stopping")
	defer log.Info().Str("stage", "shut down").Str("component", "context").Msg("stopped")
}
