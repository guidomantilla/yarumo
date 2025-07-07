package boot

import (
	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"

	"github.com/guidomantilla/yarumo/pkg/common/comm"
	clog "github.com/guidomantilla/yarumo/pkg/common/log"
	"github.com/guidomantilla/yarumo/pkg/common/pointer"
	"github.com/guidomantilla/yarumo/pkg/common/resilience"
	"github.com/guidomantilla/yarumo/pkg/common/uids"
	"github.com/guidomantilla/yarumo/pkg/security/cryptos"
	"github.com/guidomantilla/yarumo/pkg/security/hashes"
	"github.com/guidomantilla/yarumo/pkg/security/passwords"
	"github.com/guidomantilla/yarumo/pkg/security/tokens"
)

type ContainerOptions struct {
	AppName                string
	AppVersion             string
	Config                 any
	Logger                 zerolog.Logger
	Hasher                 hashes.HashFn
	UIDGen                 uids.UIDFn
	Validator              *validator.Validate
	PasswordEncoder        passwords.Encoder
	PasswordGenerator      passwords.Generator
	PasswordManager        passwords.Manager
	TokenGenerator         tokens.Generator
	Cipher                 cryptos.Cipher
	RateLimiterRegistry    *resilience.RateLimiterRegistry
	CircuitBreakerRegistry *resilience.CircuitBreakerRegistry
	HttpClient             comm.HTTPClient
	more                   map[string]any
}

func NewContainerOptions[C any](name string, version string, opts ...ContainerOption) *ContainerOptions {
	options := &ContainerOptions{
		AppName:                name,
		AppVersion:             version,
		Hasher:                 hashes.BLAKE2b_512,
		UIDGen:                 uids.UUIDv7,
		Logger:                 clog.Configure(name, version),
		Config:                 pointer.Zero[C](),
		Validator:              validator.New(),
		PasswordEncoder:        passwords.NewBcryptEncoder(),
		PasswordGenerator:      passwords.NewGenerator(),
		PasswordManager:        nil,
		TokenGenerator:         tokens.NewJwtGenerator(tokens.WithJwtIssuer(name)),
		Cipher:                 cryptos.NewAesCipher(),
		RateLimiterRegistry:    resilience.NewRateLimiterRegistry(),
		CircuitBreakerRegistry: resilience.NewCircuitBreakerRegistry(),
		HttpClient:             comm.NewHTTPClient(),
		more:                   make(map[string]any),
	}

	options.PasswordManager = passwords.NewManager(options.PasswordEncoder, options.PasswordGenerator)

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type ContainerOption func(opts *ContainerOptions)
