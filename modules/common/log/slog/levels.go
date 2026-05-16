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

// toSlog maps a Level onto the equivalent stdlib slog.Level value.
func (l Level) toSlog() slog.Level {
	return slog.Level(l)
}
