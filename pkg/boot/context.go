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
	"github.com/guidomantilla/yarumo/pkg/common/uids"
	"github.com/guidomantilla/yarumo/pkg/common/utils"
	"github.com/guidomantilla/yarumo/pkg/security/cryptos"
	"github.com/guidomantilla/yarumo/pkg/security/hashes"
	"github.com/guidomantilla/yarumo/pkg/security/passwords"
	"github.com/guidomantilla/yarumo/pkg/security/tokens"
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
	assert.NotEmpty(name, fmt.Sprintf("%s - error creating: appName is empty", "context"))
	assert.NotEmpty(version, fmt.Sprintf("%s - error creating: appName is empty", "context"))

	container := &Container{
		AppName:           name,
		AppVersion:        version,
		Hasher:            hashes.BLAKE2b_512,
		UIDGen:            uids.UUIDv7,
		Logger:            clog.Configure(name, version),
		Config:            pointer.Zero[C](),
		Validator:         validator.New(),
		PasswordEncoder:   passwords.NewBcryptEncoder(),
		PasswordGenerator: passwords.NewGenerator(),
		TokenGenerator:    tokens.NewJwtGenerator(),
		Cipher:            cryptos.NewAesCipher(),
	}

	viper.AutomaticEnv()
	options := NewOptions(opts...)

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
