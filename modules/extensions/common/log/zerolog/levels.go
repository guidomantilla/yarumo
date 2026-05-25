package zerolog

import (
	"github.com/rs/zerolog"
)

// Level represents a logging severity level for the zerolog-backed Logger.
// Values are deliberately ordered low-to-high (trace lowest, off highest)
// to make level filtering straightforward.
type Level int

// Log severity levels.
const (
	// LevelTrace enables trace-level records and everything more severe.
	LevelTrace Level = iota
	// LevelDebug enables debug-level records and everything more severe.
	LevelDebug
	// LevelInfo enables info-level records and everything more severe.
	LevelInfo
	// LevelWarn enables warn-level records and everything more severe.
	LevelWarn
	// LevelError enables error-level records and everything more severe.
	LevelError
	// LevelFatal enables fatal-level records only.
	LevelFatal
	// LevelOff disables every level. It is the default minimum level so
	// that a Logger built without explicit configuration is silent.
	LevelOff
)

// toZerolog maps a Level onto the equivalent zerolog.Level value. LevelOff
// maps to zerolog.Disabled so the underlying logger short-circuits at the
// level filter.
func (l Level) toZerolog() zerolog.Level {
	switch l {
	case LevelTrace:
		return zerolog.TraceLevel
	case LevelDebug:
		return zerolog.DebugLevel
	case LevelInfo:
		return zerolog.InfoLevel
	case LevelWarn:
		return zerolog.WarnLevel
	case LevelError:
		return zerolog.ErrorLevel
	case LevelFatal:
		return zerolog.FatalLevel
	case LevelOff:
		return zerolog.Disabled
	default:
		return zerolog.Disabled
	}
}
