// Package cron provides a scheduling abstraction over the robfig/cron library.
package cron

import (
	"context"
	"time"

	cron "github.com/robfig/cron/v3"
)

var _ Scheduler = (*cron.Cron)(nil)

var _ cron.Job = (*cron.FuncJob)(nil)

// Scheduler defines the interface for a cron job scheduler.
// The caller must call Stop to release resources when the scheduler is no longer needed.
// Implementations must be safe for concurrent use.
type Scheduler interface {
	// Start begins the cron scheduler in a separate goroutine.
	Start()
	// Run starts the cron scheduler and blocks until stopped.
	Run()
	// Stop halts the cron scheduler and returns a context that completes when running jobs finish.
	Stop() context.Context
	// AddFunc registers a function to run on the given cron schedule.
	AddFunc(spec string, cmd func()) (cron.EntryID, error)
	// AddJob registers a Job to run on the given cron schedule.
	AddJob(spec string, cmd cron.Job) (cron.EntryID, error)
	// Schedule registers a Job to run on the given pre-parsed schedule.
	Schedule(schedule cron.Schedule, cmd cron.Job) cron.EntryID
	// Entries returns a snapshot of all registered cron entries.
	Entries() []cron.Entry
	// Location returns the time zone used by the scheduler.
	Location() *time.Location
	// Entry returns the entry for the given ID.
	Entry(id cron.EntryID) cron.Entry
	// Remove cancels and removes the entry with the given ID.
	Remove(id cron.EntryID)
}
