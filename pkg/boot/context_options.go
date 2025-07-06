package boot

import "github.com/guidomantilla/yarumo/pkg/common/utils"

type WireContextOptions struct {
	Hasher            BeanFn
	UIDGen            BeanFn
	Logger            BeanFn
	Config            BeanFn
	Validator         BeanFn
	PasswordEncoder   BeanFn
	PasswordGenerator BeanFn
	TokenGenerator    BeanFn
	Cipher            BeanFn
	HttpClient        BeanFn
	More              []BeanFn
}

func NewOptions(opts ...WireContextOption) *WireContextOptions {
	options := &WireContextOptions{
		Hasher:            Hasher,
		UIDGen:            UIDGen,
		Logger:            Logger,
		Config:            Config,
		Validator:         Validator,
		PasswordEncoder:   PasswordEncoder,
		PasswordGenerator: PasswordGenerator,
		TokenGenerator:    TokenGenerator,
		Cipher:            Cipher,
		HttpClient:        HttpClient,
		More:              make([]BeanFn, 0),
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

// WithLogger allows setting a custom logger function into the WireContext (wctx *boot.WireContext).
//
// wctx.Logger = <logger object>
func WithLogger(loggerFn BeanFn) WireContextOption {
	return func(opts *WireContextOptions) {
		if utils.NotNil(loggerFn) {
			opts.Logger = loggerFn
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
