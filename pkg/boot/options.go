package boot

import (
	clog "github.com/guidomantilla/yarumo/pkg/common/log"
	"github.com/guidomantilla/yarumo/pkg/common/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/guidomantilla/yarumo/pkg/common/pointer"
)

type Options struct {
	Config ConfigFn
}

func NewOptions(opts ...Option) *Options {
	options := &Options{
		Config: func(wctx *WireContext) any {
			log.Warn().Str("stage", "startup").Str("component", "configuration").Msg("config function not implemented. using default configuration")
			viper.AutomaticEnv()
			debugMode := utils.Ternary(viper.IsSet("DEBUG_MODE"), viper.GetBool("DEBUG_MODE"), false)
			clogOpts := clog.Chain().
				WithCaller(debugMode).
				WithGlobalLevel(utils.Ternary(debugMode, zerolog.DebugLevel, wctx.LogLevel)).
				Build()
			clog.Configure(wctx.AppName, wctx.AppVersion, clogOpts)
			return pointer.Zero[any]()
		},
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type Option func(opts *Options)

func WithConfig(configFn ConfigFn) Option {
	return func(opts *Options) {
		opts.Config = configFn
	}
}

//

type OptionsChain struct {
	chain []Option
}

func Chain() *OptionsChain {
	return &OptionsChain{
		chain: make([]Option, 0),
	}
}

func (chain *OptionsChain) Build() Option {
	return func(opts *Options) {
		for _, option := range chain.chain {
			option(opts)
		}
	}
}

func (chain *OptionsChain) WithConfig(configFn ConfigFn) *OptionsChain {
	chain.chain = append(chain.chain, WithConfig(configFn))
	return chain
}
