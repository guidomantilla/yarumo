package cron

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler interface {
	Start()
	Run()
	Stop() context.Context
	AddFunc(spec string, cmd func()) (cron.EntryID, error)
	AddJob(spec string, cmd cron.Job) (cron.EntryID, error)
	Schedule(schedule cron.Schedule, cmd cron.Job) cron.EntryID
	Entries() []cron.Entry
	Location() *time.Location
	Entry(id cron.EntryID) cron.Entry
	Remove(id cron.EntryID)
}
