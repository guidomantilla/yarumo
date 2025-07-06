package core

type Config struct {
	DebugMode    bool   `mapstructure:"DEBUG_MODE"`
	Host         string `mapstructure:"HOST"`
	HttpPort     string `mapstructure:"HTTP_PORT"`
	GrpcPort     string `mapstructure:"GRPC_PORT"`
	CipherKey    string `mapstructure:"CIPHER_KEY"`
	TokenKey     string `mapstructure:"TOKEN_KEY"`
	TokenTimeout string `mapstructure:"TOKEN_TIMEOUT"`
}
