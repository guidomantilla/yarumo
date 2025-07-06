package boot

import (
	"github.com/go-playground/validator/v10"
	"github.com/guidomantilla/yarumo/pkg/common/comm"
	clog "github.com/guidomantilla/yarumo/pkg/common/log"
	"github.com/guidomantilla/yarumo/pkg/common/pointer"
	"github.com/guidomantilla/yarumo/pkg/common/uids"
	"github.com/guidomantilla/yarumo/pkg/security/cryptos"
	"github.com/guidomantilla/yarumo/pkg/security/hashes"
	"github.com/guidomantilla/yarumo/pkg/security/passwords"
	"github.com/guidomantilla/yarumo/pkg/security/tokens"
	"github.com/rs/zerolog"
)

type ContainerOptions struct {
	AppName           string
	AppVersion        string
	Config            any
	Logger            zerolog.Logger
	Hasher            hashes.HashFn
	UIDGen            uids.UIDFn
	Validator         *validator.Validate
	PasswordEncoder   passwords.Encoder
	PasswordGenerator passwords.Generator
	TokenGenerator    tokens.Generator
	Cipher            cryptos.Cipher
	HttpClient        comm.HTTPClient
	more              map[string]any
}

func NewContainerOptions[C any](opts ...ContainerOption) *ContainerOptions {
	options := &ContainerOptions{
		AppName:           "",
		AppVersion:        "",
		Hasher:            hashes.BLAKE2b_512,
		UIDGen:            uids.UUIDv7,
		Logger:            clog.Configure("", ""),
		Config:            pointer.Zero[C](),
		Validator:         validator.New(),
		PasswordEncoder:   passwords.NewBcryptEncoder(),
		PasswordGenerator: passwords.NewGenerator(),
		TokenGenerator:    tokens.NewJwtGenerator(),
		Cipher:            cryptos.NewAesCipher(),
		HttpClient:        comm.NewHTTPClient(),
		more:              make(map[string]any),
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type ContainerOption func(opts *ContainerOptions)
