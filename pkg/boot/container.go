package boot

import (
	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"

	"github.com/guidomantilla/yarumo/pkg/common/comm"
	"github.com/guidomantilla/yarumo/pkg/common/pointer"
	"github.com/guidomantilla/yarumo/pkg/common/resilience"
	"github.com/guidomantilla/yarumo/pkg/common/uids"
	"github.com/guidomantilla/yarumo/pkg/security/cryptos"
	"github.com/guidomantilla/yarumo/pkg/security/hashes"
	"github.com/guidomantilla/yarumo/pkg/security/passwords"
	"github.com/guidomantilla/yarumo/pkg/security/tokens"
)

type Container struct {
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

func NewContainer[C any](name string, version string, opts ...ContainerOption) *Container {
	options := NewContainerOptions[C](name, version, opts...)
	return &Container{
		AppName:                options.AppName,
		AppVersion:             options.AppVersion,
		Config:                 options.Config,
		Logger:                 options.Logger,
		Hasher:                 options.Hasher,
		UIDGen:                 options.UIDGen,
		Validator:              options.Validator,
		PasswordEncoder:        options.PasswordEncoder,
		PasswordGenerator:      options.PasswordGenerator,
		PasswordManager:        options.PasswordManager,
		TokenGenerator:         options.TokenGenerator,
		Cipher:                 options.Cipher,
		RateLimiterRegistry:    options.RateLimiterRegistry,
		CircuitBreakerRegistry: options.CircuitBreakerRegistry,
		HttpClient:             options.HttpClient,
		more:                   options.more,
	}
}

func Add(container *Container, key string, value any) {
	container.more[key] = value
}

func Get[T any](container *Container, key string) T {
	value, exists := container.more[key]
	if !exists {
		return pointer.Zero[T]()
	}

	typedValue, ok := value.(T)
	if !ok {
		return pointer.Zero[T]()
	}

	return typedValue
}
