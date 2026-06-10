package cron

import (
	"context"

	"github.com/robfig/cron/v3"

	clog "github.com/guidomantilla/yarumo/core/common/log"
)

type cronlogger struct {
	ctx    context.Context
	logger clog.Logger
}

func WrapLogger(ctx context.Context, logger clog.Logger) cron.Logger {
	if ctx == nil {
		ctx = context.Background()
	}
	return &cronlogger{ctx: ctx, logger: logger}
}

func (l cronlogger) Info(msg string, keysAndValues ...any) {
	l.logger.Info(l.ctx, msg, keysAndValues...)
}

func (l cronlogger) Error(err error, msg string, keysAndValues ...any) {
	args := append([]any{"error", err}, keysAndValues...)
	l.logger.Error(l.ctx, msg, args...)
}

func WithLogger(ctx context.Context, logger clog.Logger) cron.Option {
	return cron.WithLogger(WrapLogger(ctx, logger))
}
