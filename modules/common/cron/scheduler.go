package cron

import "github.com/robfig/cron/v3"

func NewScheduler(options ...cron.Option) Scheduler {
	return cron.New(options...)
}
