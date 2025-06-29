package environment

import (
	"io"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Options struct {
}

func NewOptions(opts ...Option) *Options {
	options := &Options{}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type Option func(options *Options)

func WithConfigName(name string) Option {
	return func(_ *Options) {
		viper.SetConfigName(name)
	}
}

func WithConfigType(configType string) Option {
	return func(_ *Options) {
		viper.SetConfigType(configType)
	}
}

func WithConfigPath(path string) Option {
	return func(_ *Options) {
		viper.AddConfigPath(path)
	}
}

func WithConfigFile(file string) Option {
	return func(_ *Options) {
		viper.SetConfigFile(file)
	}
}

func WithConfig(reader io.Reader) Option {
	return func(_ *Options) {
		err := viper.MergeConfig(reader)
		if err != nil {
			log.Error().Str("config_file", viper.ConfigFileUsed()).Err(err).Msg("failed to merge config file")
		}
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
	return func(options *Options) {
		for _, option := range chain.chain {
			option(options)
		}
	}
}

func (chain *OptionsChain) WithConfigName(name string) *OptionsChain {
	chain.chain = append(chain.chain, WithConfigName(name))
	return chain
}

func (chain *OptionsChain) WithConfigType(configType string) *OptionsChain {
	chain.chain = append(chain.chain, WithConfigType(configType))
	return chain
}

func (chain *OptionsChain) WithConfigPath(path string) *OptionsChain {
	chain.chain = append(chain.chain, WithConfigPath(path))
	return chain
}

func (chain *OptionsChain) WithConfigFile(file string) *OptionsChain {
	chain.chain = append(chain.chain, WithConfigFile(file))
	return chain
}

func (chain *OptionsChain) WithConfig(reader io.Reader) *OptionsChain {
	chain.chain = append(chain.chain, WithConfig(reader))
	return chain
}
