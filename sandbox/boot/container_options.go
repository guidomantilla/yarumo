package boot

import (
	"fmt"

	validator "github.com/go-playground/validator/v10"

	"github.com/guidomantilla/yarumo/modules/common/assert"
	"github.com/guidomantilla/yarumo/modules/common/pointer"
	"github.com/guidomantilla/yarumo/pkg/security/cryptos"
	"github.com/guidomantilla/yarumo/pkg/security/hashes"
	"github.com/guidomantilla/yarumo/pkg/security/passwords"
	"github.com/guidomantilla/yarumo/pkg/security/tokens"
	"github.com/guidomantilla/yarumo/pkg/uids"
	comm "github.com/guidomantilla/yarumo/sandbox/comm"
	resilience "github.com/guidomantilla/yarumo/sandbox/resilience"
)

type ContainerOptions struct {
	appName                string
	appVersion             string
	config                 any
	hasher                 hashes.HashFn
	uidGen                 uids.UIDFn
	validator              *validator.Validate
	passwordEncoder        passwords.Encoder
	passwordGenerator      passwords.Generator
	passwordManager        passwords.Manager
	tokenGenerator         tokens.Generator
	cipher                 cryptos.Cipher
	rateLimiterRegistry    *resilience.RateLimiterRegistry
	circuitBreakerRegistry *resilience.CircuitBreakerRegistry
	httpClient             comm.HTTPClient
	more                   map[string]any
}

func NewContainerOptions[C any](name string, version string, opts ...ContainerOption) *ContainerOptions {
	assert.NotEmpty(name, fmt.Sprintf("%s - error starting up: name is nil", "container-options"))
	assert.NotEmpty(version, fmt.Sprintf("%s - error starting up: version is nil", "container-options"))
	options := &ContainerOptions{
		appName:                name,
		appVersion:             version,
		hasher:                 hashes.BLAKE2b_512,
		uidGen:                 uids.UUIDv7,
		config:                 pointer.Zero[C](),
		validator:              validator.New(),
		passwordEncoder:        passwords.NewBcryptEncoder(),
		passwordGenerator:      passwords.NewGenerator(),
		passwordManager:        nil,
		tokenGenerator:         tokens.NewJwtGenerator(tokens.WithJwtIssuer(name)),
		cipher:                 cryptos.NewAesCipher(),
		rateLimiterRegistry:    resilience.NewRateLimiterRegistry(),
		circuitBreakerRegistry: resilience.NewCircuitBreakerRegistry(),
		httpClient:             comm.NewHTTPClient(),
		more:                   make(map[string]any),
	}

	options.passwordManager = passwords.NewManager(options.passwordEncoder, options.passwordGenerator)

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type ContainerOption func(opts *ContainerOptions)
