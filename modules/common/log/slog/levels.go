// Package slog provides a structured logger implementation built on top of log/slog.
package slog

import "log/slog"

// Level represents a logging severity level.
type Level int

// Log severity levels.
const (
	LevelTrace Level = -8
	LevelDebug Level = -4
	LevelInfo  Level = 0
	LevelWarn  Level = 4
	LevelError Level = 8
	LevelFatal Level = 12
	LevelOff   Level = 16
)

func (l Level) toSlog() slog.Level {
	return slog.Level(l)
}

// ReplaceLevel replaces the level attribute with a more readable value.
func ReplaceLevel(_ []string, a slog.Attr) slog.Attr {
	if a.Key != slog.LevelKey {
		return a
	}

	level, ok := a.Value.Any().(slog.Level)
	if !ok {
		return a
	}

	switch Level(level) {
	case LevelTrace:
		a.Value = slog.StringValue("TRACE")
	case LevelFatal:
		a.Value = slog.StringValue("FATAL")
	default:
	}

	return a
}
