package zerolog

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

// emit applies the variadic key-value args to the zerolog event and writes
// the message. Args are interpreted as alternating key/value pairs; an odd
// trailing arg is recorded under the conventional zerolog "!BADKEY" field
// so the caller does not silently lose data. Nil events (zerolog returns
// nil when the level is disabled) are skipped.
func emit(event *zerolog.Event, ctx context.Context, msg string, args ...any) {
	if event == nil {
		return
	}

	if ctx != nil {
		ctxErr := ctx.Err()
		if ctxErr != nil {
			event = event.AnErr("ctx_err", ctxErr)
		}
	}

	event = applyArgs(event, args)
	event.Msg(msg)
}

// applyArgs walks args as alternating key/value pairs and adds each pair
// to the event using the most specific zerolog typed setter available. An
// odd trailing arg is attached under the "!BADKEY" field so it remains
// visible in the output rather than being silently dropped.
func applyArgs(event *zerolog.Event, args []any) *zerolog.Event {
	if len(args) == 0 {
		return event
	}

	i := 0
	for i < len(args) {
		if i == len(args)-1 {
			event = event.Interface("!BADKEY", args[i])

			break
		}

		key, ok := args[i].(string)
		if !ok {
			key = fmt.Sprintf("!BADKEY_%d", i)
		}

		event = appendField(event, key, args[i+1])
		i += 2
	}

	return event
}

// appendField adds a single typed field to the event, choosing the most
// specific zerolog setter for the value's runtime type. Unknown types fall
// back to .Interface which serialises with reflection.
//
//nolint:cyclop // exhaustive runtime-type switch is simpler than a map dispatch here.
func appendField(event *zerolog.Event, key string, value any) *zerolog.Event {
	switch v := value.(type) {
	case nil:
		return event.Interface(key, nil)
	case string:
		return event.Str(key, v)
	case bool:
		return event.Bool(key, v)
	case int:
		return event.Int(key, v)
	case int8:
		return event.Int8(key, v)
	case int16:
		return event.Int16(key, v)
	case int32:
		return event.Int32(key, v)
	case int64:
		return event.Int64(key, v)
	case uint:
		return event.Uint(key, v)
	case uint8:
		return event.Uint8(key, v)
	case uint16:
		return event.Uint16(key, v)
	case uint32:
		return event.Uint32(key, v)
	case uint64:
		return event.Uint64(key, v)
	case float32:
		return event.Float32(key, v)
	case float64:
		return event.Float64(key, v)
	case time.Time:
		return event.Time(key, v)
	case time.Duration:
		return event.Dur(key, v)
	case error:
		return event.AnErr(key, v)
	default:
		return event.Interface(key, v)
	}
}
