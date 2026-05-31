package claimcheck

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/guidomantilla/yarumo/messaging"
)

// defaultDeleteAfterRetrieve is the default for WithDeleteAfterRetrieve
// on ClaimCheckOut — true so a single-consumer claim does not leak
// store entries. Multi-consumer or replay scenarios should pass
// WithDeleteAfterRetrieve(false) and manage cleanup out-of-band.
const defaultDeleteAfterRetrieve = true

// keyByteLength is the number of random bytes drawn by defaultKeyGen
// before hex-encoding. 16 bytes (128 bits) gives a collision-resistant
// opaque key without paying for UUID parsing/serialization.
const keyByteLength = 16

// Option is a functional option for configuring claim check Options.
// Both ClaimCheckIn and ClaimCheckOut share the same Option type
// because the option set is small and consumer-side overrides (Out)
// silently no-op on producer-side options (In) and vice versa — the
// alternative of two separate Option types adds API surface without
// catching a real bug class.
type Option func(opts *Options)

// Options holds the configuration for ClaimCheckIn and ClaimCheckOut.
type Options struct {
	keyGen                KeyGenFn
	deleteAfterRetrieve   bool
	errorHandler          messaging.ErrorHandler
}

// NewOptions creates a new Options with sensible defaults and applies
// the given options. Defaults:
//
//   - KeyGen: defaultKeyGen (16 random bytes hex-encoded).
//   - DeleteAfterRetrieve: true (Out deletes after a successful Get).
//   - ErrorHandler: messaging.DefaultErrorHandler (logs via common/log).
func NewOptions(opts ...Option) *Options {
	options := &Options{
		keyGen:              defaultKeyGen,
		deleteAfterRetrieve: defaultDeleteAfterRetrieve,
		errorHandler:        messaging.DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithKeyGen installs the key generator used by ClaimCheckIn for each
// stored message. The default (defaultKeyGen) draws 16 bytes from
// crypto/rand and hex-encodes them. Replace this when the store
// backend wants a specific key shape (ULID, broker-assigned id, …) or
// in tests to use a deterministic counter. Nil values are ignored
// (the previously installed generator is preserved). ClaimCheckOut
// silently ignores this option — it never generates keys.
func WithKeyGen(fn KeyGenFn) Option {
	return func(opts *Options) {
		if fn != nil {
			opts.keyGen = fn
		}
	}
}

// WithDeleteAfterRetrieve configures whether ClaimCheckOut deletes the
// retrieved entry from the MessageStore after a successful Get. The
// default is true — single-consumer claims do not leak store entries.
// Pass WithDeleteAfterRetrieve(false) for multi-consumer or replay
// scenarios where the same key may be retrieved more than once;
// cleanup is then the caller's responsibility. ClaimCheckIn silently
// ignores this option — it never deletes.
func WithDeleteAfterRetrieve(delete bool) Option {
	return func(opts *Options) {
		opts.deleteAfterRetrieve = delete
	}
}

// WithErrorHandler installs an observability hook fired once per real
// claim check failure (store Put/Get/Delete failed, forward Send
// failed). The default (when WithErrorHandler is not passed) is
// messaging.DefaultErrorHandler, which logs each failure via
// common/log so consumers that forget to wire observability still see
// real failures. Pass messaging.SilentErrorHandler to opt out, or any
// custom hook to redirect. Nil values are ignored (the previously
// installed handler is preserved).
func WithErrorHandler(handler messaging.ErrorHandler) Option {
	return func(opts *Options) {
		if handler != nil {
			opts.errorHandler = handler
		}
	}
}

// defaultKeyGen returns a 128-bit hex-encoded random string drawn from
// crypto/rand. If crypto/rand fails (extraordinarily rare in
// production but possible in entropy-starved test environments), the
// function returns an empty string. ClaimCheckIn treats an empty key
// as a Put failure surface only when MessageStore.Put rejects it —
// most stores accept empty keys, which would still produce a working
// claim because the same generator is used for the matching
// reference. Consumers that need defensive validation should wire a
// WithKeyGen that retries or panics on failure.
func defaultKeyGen() string {
	buf := make([]byte, keyByteLength)

	_, err := rand.Read(buf)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(buf)
}
