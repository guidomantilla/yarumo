package boot

import (
	"fmt"
	"io"

	validator "github.com/go-playground/validator/v10"

	"github.com/guidomantilla/yarumo/modules/common/assert"
	"github.com/guidomantilla/yarumo/modules/common/pointer"
	"github.com/guidomantilla/yarumo/pkg/comm"
	resilience "github.com/guidomantilla/yarumo/pkg/resilience"
	"github.com/guidomantilla/yarumo/pkg/security/cryptos"
	"github.com/guidomantilla/yarumo/pkg/security/hashes"
	"github.com/guidomantilla/yarumo/pkg/security/passwords"
	"github.com/guidomantilla/yarumo/pkg/security/tokens"
	"github.com/guidomantilla/yarumo/pkg/uids"
)

type Container struct {
	AppName                string
	AppVersion             string
	LoggerWriters          []io.Writer
	Config                 any
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
	assert.NotEmpty(name, fmt.Sprintf("%s - error starting up: name is nil", "container"))
	assert.NotEmpty(version, fmt.Sprintf("%s - error starting up: version is nil", "container"))
	options := NewContainerOptions[C](name, version, opts...)
	return &Container{
		AppName:                options.appName,
		AppVersion:             options.appVersion,
		Config:                 options.config,
		Hasher:                 options.hasher,
		UIDGen:                 options.uidGen,
		Validator:              options.validator,
		PasswordEncoder:        options.passwordEncoder,
		PasswordGenerator:      options.passwordGenerator,
		PasswordManager:        options.passwordManager,
		TokenGenerator:         options.tokenGenerator,
		Cipher:                 options.cipher,
		RateLimiterRegistry:    options.rateLimiterRegistry,
		CircuitBreakerRegistry: options.circuitBreakerRegistry,
		HttpClient:             options.httpClient,
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
