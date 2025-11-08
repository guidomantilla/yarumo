package boot

import "github.com/guidomantilla/yarumo/modules/common/utils"

type WireContextOptions struct {
	Hasher                 BeanFn
	UIDGen                 BeanFn
	Config                 BeanFn
	Validator              BeanFn
	PasswordEncoder        BeanFn
	PasswordGenerator      BeanFn
	PasswordManager        BeanFn
	TokenGenerator         BeanFn
	Cipher                 BeanFn
	RateLimiterRegistry    BeanFn
	CircuitBreakerRegistry BeanFn
	HttpClient             BeanFn
	More                   []BeanFn
}

func NewOptions(opts ...WireContextOption) *WireContextOptions {
	options := &WireContextOptions{
		Hasher:                 Hasher,
		UIDGen:                 UIDGen,
		Config:                 Config,
		Validator:              Validator,
		PasswordEncoder:        PasswordEncoder,
		PasswordGenerator:      PasswordGenerator,
		PasswordManager:        PasswordManager,
		TokenGenerator:         TokenGenerator,
		Cipher:                 Cipher,
		RateLimiterRegistry:    RateLimiterRegistry,
		CircuitBreakerRegistry: BreakerRegistry,
		HttpClient:             HttpClient,
		More:                   make([]BeanFn, 0),
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

type WireContextOption func(opts *WireContextOptions)

// WithHasher allows setting a custom hasher function into the WireContext (wctx *boot.WireContext).
//
// wctx.Hasher = <hasher object>
func WithHasher(hasherFn BeanFn) WireContextOption {
	return func(opts *WireContextOptions) {
		if utils.NotNil(hasherFn) {
			opts.Hasher = hasherFn
		}
	}
}

// WithUIDGen allows setting a custom UID generator function into the WireContext (wctx *boot.WireContext).
//
// wctx.UIDGen = <uid generator object>
func WithUIDGen(uidGenFn BeanFn) WireContextOption {
	return func(opts *WireContextOptions) {
		if utils.NotNil(uidGenFn) {
			opts.UIDGen = uidGenFn
		}
	}
}

// WithConfig allows setting a custom config function into the WireContext (wctx *boot.WireContext).
//
// wctx.Config = <config object>
func WithConfig(configFn BeanFn) WireContextOption {
	return func(opts *WireContextOptions) {
		if utils.NotNil(configFn) {
			opts.Config = configFn
		}
	}
}

// WithValidator allows setting a custom validator function into the WireContext (wctx *boot.WireContext).
//
// wctx.Validator = <validator object>
func WithValidator(validatorFn BeanFn) WireContextOption {
	return func(opts *WireContextOptions) {
		if utils.NotNil(validatorFn) {
			opts.Validator = validatorFn
		}
	}
}

// WithPasswordEncoder allows setting a custom password encoder function into the WireContext (wctx *boot.WireContext).
//
// wctx.PasswordEncoder = <password encoder object>
func WithPasswordEncoder(passwordEncoderFn BeanFn) WireContextOption {
	return func(opts *WireContextOptions) {
		if utils.NotNil(passwordEncoderFn) {
			opts.PasswordEncoder = passwordEncoderFn
		}
	}
}

// WithPasswordGenerator allows setting a custom password generator function into the WireContext (wctx *boot.WireContext).
//
// wctx.PasswordGenerator = <password generator object>
func WithPasswordGenerator(passwordGeneratorFn BeanFn) WireContextOption {
	return func(opts *WireContextOptions) {
		if utils.NotNil(passwordGeneratorFn) {
			opts.PasswordGenerator = passwordGeneratorFn
		}
	}
}

// WithPasswordManager allows setting a custom password manager function into the WireContext (wctx *boot.WireContext).
// This function is a combination of password encoder and password generator.
//
// wctx.PasswordManager = <password manager object>
func WithPasswordManager(passwordManagerFn BeanFn) WireContextOption {
	return func(opts *WireContextOptions) {
		if utils.NotNil(passwordManagerFn) {
			opts.PasswordManager = passwordManagerFn
		}
	}
}

// WithTokenGenerator allows setting a custom token generator function into the WireContext (wctx *boot.WireContext).
//
// wctx.TokenGenerator = <token generator object>
func WithTokenGenerator(tokenGeneratorFn BeanFn) WireContextOption {
	return func(opts *WireContextOptions) {
		if utils.NotNil(tokenGeneratorFn) {
			opts.TokenGenerator = tokenGeneratorFn
		}
	}
}

// WithCipher allows setting a custom cipher function into the WireContext (wctx *boot.WireContext).
//
// wctx.Cipher = <cipher object>
func WithCipher(cipherFn BeanFn) WireContextOption {
	return func(opts *WireContextOptions) {
		if utils.NotNil(cipherFn) {
			opts.Cipher = cipherFn
		}
	}
}

// WithRateLimiterRegistry allows setting a custom rate limiter registry function into the WireContext (wctx *boot.WireContext).
//
// wctx.RateLimiterRegistry = <rate limiter registry object>
func WithRateLimiterRegistry(rateLimiterRegistryFn BeanFn) WireContextOption {
	return func(opts *WireContextOptions) {
		if utils.NotNil(rateLimiterRegistryFn) {
			opts.RateLimiterRegistry = rateLimiterRegistryFn
		}
	}
}

// WithCircuitBreakerRegistry allows setting a custom circuit breaker registry function into the WireContext (wctx *boot.WireContext).
//
// wctx.WithCircuitBreakerRegistry = <circuit breaker registry object>
func WithCircuitBreakerRegistry(circuitBreakerRegistryFn BeanFn) WireContextOption {
	return func(opts *WireContextOptions) {
		if utils.NotNil(circuitBreakerRegistryFn) {
			opts.CircuitBreakerRegistry = circuitBreakerRegistryFn
		}
	}
}

// WithHttpClient allows setting a custom HTTP client function into the WireContext (wctx *boot.WireContext).
//
// wctx.HttpClient = <http client object>
func WithHttpClient(httpClientFn BeanFn) WireContextOption {
	return func(opts *WireContextOptions) {
		if utils.NotNil(httpClientFn) {
			opts.HttpClient = httpClientFn
		}
	}
}

// With allows adding more custom functions into the WireContext (wctx *boot.WireContext).
// These functions will be executed after the main functions defined in the Options struct.
func With(withFn BeanFn) WireContextOption {
	return func(opts *WireContextOptions) {
		if utils.NotNil(withFn) {
			opts.More = append(opts.More, withFn)
		}
	}
}
