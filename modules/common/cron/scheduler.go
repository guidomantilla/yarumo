package cron

import cron "github.com/robfig/cron/v3"

// NewScheduler creates a new cron scheduler.
func NewScheduler(options ...cron.Option) Scheduler {
	return cron.New(options...)
}
