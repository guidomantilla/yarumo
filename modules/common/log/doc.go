// Package log provides a structured logging abstraction with support for
// multiple log levels.
//
// # Process-global default logger
//
// The package holds a process-global default logger in `current` / `internal`
// (see internals.go). The intended lifecycle is:
//
//  1. Application startup wires the concrete logger (typically
//     `slog.NewLogger(...)` from the sibling `modules/log/slog` module,
//     optionally decorated with extractors such as
//     `otelslog.WithOtelTrace()`).
//  2. Startup calls `log.Use(logger)` exactly once, before any goroutine
//     observes the default slot.
//  3. Producing code calls the package-level helpers (`Trace`, `Debug`,
//     `Info`, `Warn`, `Error`, `Fatal`); each helper resolves the current
//     logger via the atomic slot.
//
// Until Use is called, the slot serves a noopLogger that discards
// Trace/Debug/Info/Warn/Error and exits the process on Fatal (writing the
// message to stderr first) so that "no logger configured" cannot hide a
// fatal condition triggered by assert or other callers.
//
// This package intentionally keeps zero dependencies on concrete logger
// implementations. Concrete loggers live in their own modules
// (modules/log/slog, etc.) and depend on this interface, never the
// reverse: the dependency direction is consumer -> abstraction, never
// common -> implementation.
//
// Tests in this package mutate the slot via `Use` and the private `load`
// helper. They are intentionally serial (no `t.Parallel()`): running them
// concurrently would race against any other parallel test in the same
// package that observes the slot. Tests in downstream packages (and in the
// `modules/log/slog` and `modules/log/slog/slogctx` subpackages) carry no
// global state and run fully parallel.
//
// Do not introduce new tests in this package that depend on the default
// slot without ensuring they remain serial.
package log
