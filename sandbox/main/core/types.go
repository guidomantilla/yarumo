package core

type Config struct {
	DebugMode    bool   `mapstructure:"DEBUG_MODE"`
	LogLevel     string `mapstructure:"LOG_LEVEL"`
	CipherKey    string `mapstructure:"CIPHER_KEY"`
	TokenKey     string `mapstructure:"TOKEN_KEY"`
	TokenTimeout string `mapstructure:"TOKEN_TIMEOUT"`
}
