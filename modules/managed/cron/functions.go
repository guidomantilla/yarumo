package cron

import (
	"context"

	"github.com/robfig/cron/v3"

	clog "github.com/guidomantilla/yarumo/core/common/log"
)

func WithLogger(ctx context.Context, logger clog.Logger) cron.Option {
	return cron.WithLogger(WrapLogger(ctx, logger))
}

func WithRecoverableSkipIfStillRunning(ctx context.Context, logger clog.Logger) cron.Option {
	return cron.WithChain(cron.SkipIfStillRunning(WrapLogger(ctx, logger)), cron.Recover(WrapLogger(ctx, logger)))
}
