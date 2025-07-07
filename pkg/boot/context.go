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
	container := NewContainer[C](name, version)
	logger := log.With().Str("stage", "startup").Str("component", "context").Logger()

	logger.Info().Msg("starting")
	defer logger.Info().Msg("started")

	options.Logger(container)
	logger.Info().Msg("logger set up")

	options.Hasher(container)
	logger.Info().Msg("hasher set up")

	options.UIDGen(container)
	logger.Info().Msg("uid generator set up")

	options.Config(container)
	logger.Info().Msg("configuration set up")

	options.Validator(container)
	logger.Info().Msg("validator set up")

	options.PasswordEncoder(container)
	logger.Info().Msg("password encoder set up")

	options.PasswordGenerator(container)
	logger.Info().Msg("password generator set up")

	options.PasswordManager(container)
	logger.Info().Msg("password manager set up")

	options.TokenGenerator(container)
	logger.Info().Msg("token generator set up")

	options.Cipher(container)
	logger.Info().Msg("cipher set up")

	options.RateLimiterRegistry(container)
	logger.Info().Msg("rate limiter registry set up")

	options.CircuitBreakerRegistry(container)
	logger.Info().Msg("circuit breaker registry set up")

	options.HttpClient(container)
	logger.Info().Msg("http client set up")

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
	logger := log.With().Str("stage", "shut down").Str("component", "context").Logger()

	logger.Info().Msg("stopping")
	defer logger.Info().Msg("stopped")
}
