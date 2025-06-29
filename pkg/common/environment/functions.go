package environment

import "github.com/spf13/viper"

func Configure(opts ...Option) {
	viper.AutomaticEnv()
}
