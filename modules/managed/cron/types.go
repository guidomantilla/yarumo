// Package cron provides a scheduling abstraction over the robfig/cron library.
//
// Scheduler is created via NewScheduler with a name and optional cron.Options.
// It implements common/lifecycle.Component (Name + Start + Stop + Done) with
// worker-style semantics: Start kicks off the cron's internal goroutines and
// returns immediately; Done closes after Stop completes.
//
// Consumers wire the Scheduler into the lifecycle pipeline via
// lifecycle.Build(ctx, scheduler, errChan), which returns the CloseFn for
// graceful shutdown.
//
// Concurrency: Scheduler implementations are safe for concurrent use by
// multiple goroutines.
package cron

import (
	"time"

	cron "github.com/robfig/cron/v3"

	"github.com/guidomantilla/yarumo/core/common/lifecycle"
)

var (
	_ Scheduler = (*scheduler)(nil)
)

// Scheduler defines the interface for a cron job scheduler.
//
// The caller must call Stop to release resources when the scheduler is no
// longer needed. Implementations must be safe for concurrent use by multiple
// goroutines.
type Scheduler interface {
	lifecycle.Component
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
