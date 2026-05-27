// Package main demonstrates the common/health primitives: a handful of
// stub Check implementations (mock DB, mock cache, mock external API, mock
// disk) aggregated through a single synchronous Health invocation.
//
// The example is intentionally self-contained — no real network or storage
// is touched — so it can be run from the repository root with
//
//	go run ./modules/common/health/examples
package main

import (
	"context"
	"fmt"
	"time"

	chealth "github.com/guidomantilla/yarumo/core/common/health"
)

// dbCheck is a stub health probe that simulates a database ping.
type dbCheck struct {
	name    string
	healthy bool
	latency time.Duration
}

// Name returns the check name.
func (c *dbCheck) Name() string { return c.name }

// Probe simulates a database health check.
func (c *dbCheck) Probe(ctx context.Context) chealth.Result {
	select {
	case <-ctx.Done():
		return chealth.Result{Status: chealth.StatusUnhealthy, Message: "ctx cancelled"}
	case <-time.After(c.latency):
	}

	if c.healthy {
		return chealth.Result{
			Status:  chealth.StatusHealthy,
			Message: "ping ok",
			Details: map[string]any{"latencyMs": c.latency.Milliseconds()},
		}
	}

	return chealth.Result{
		Status:  chealth.StatusUnhealthy,
		Message: "ping failed",
	}
}

// cacheCheck is a stub health probe that simulates a cache hit-rate check.
type cacheCheck struct {
	name    string
	hitRate float64
}

// Name returns the check name.
func (c *cacheCheck) Name() string { return c.name }

// Probe simulates a cache health check. A hit rate below 0.5 yields
// degraded, below 0.2 yields unhealthy.
func (c *cacheCheck) Probe(_ context.Context) chealth.Result {
	if c.hitRate < 0.2 {
		return chealth.Result{
			Status:  chealth.StatusUnhealthy,
			Message: "cache effectively cold",
			Details: map[string]any{"hitRate": c.hitRate},
		}
	}

	if c.hitRate < 0.5 {
		return chealth.Result{
			Status:  chealth.StatusDegraded,
			Message: "low cache hit rate",
			Details: map[string]any{"hitRate": c.hitRate},
		}
	}

	return chealth.Result{
		Status:  chealth.StatusHealthy,
		Message: "cache warm",
		Details: map[string]any{"hitRate": c.hitRate},
	}
}

// apiCheck is a stub health probe that simulates an external API ping.
type apiCheck struct {
	name        string
	statusCode  int
	probeJitter time.Duration
}

// Name returns the check name.
func (c *apiCheck) Name() string { return c.name }

// Probe simulates an external API health check.
func (c *apiCheck) Probe(ctx context.Context) chealth.Result {
	select {
	case <-ctx.Done():
		return chealth.Result{Status: chealth.StatusUnhealthy, Message: "ctx cancelled"}
	case <-time.After(c.probeJitter):
	}

	if c.statusCode >= 500 {
		return chealth.Result{Status: chealth.StatusUnhealthy, Message: "upstream 5xx"}
	}

	if c.statusCode >= 400 {
		return chealth.Result{Status: chealth.StatusDegraded, Message: "upstream 4xx"}
	}

	return chealth.Result{Status: chealth.StatusHealthy, Message: "upstream ok"}
}

// diskCheck is a stub health probe that simulates a disk-space check.
type diskCheck struct {
	name      string
	freeRatio float64
}

// Name returns the check name.
func (c *diskCheck) Name() string { return c.name }

// Probe simulates a disk-space health check.
func (c *diskCheck) Probe(_ context.Context) chealth.Result {
	if c.freeRatio < 0.05 {
		return chealth.Result{Status: chealth.StatusUnhealthy, Message: "disk full"}
	}

	if c.freeRatio < 0.2 {
		return chealth.Result{Status: chealth.StatusDegraded, Message: "disk low"}
	}

	return chealth.Result{Status: chealth.StatusHealthy, Message: "disk ok"}
}

func main() {
	h := chealth.NewHealth(chealth.WithConcurrency(4))

	h.Register(&dbCheck{name: "primary-db", healthy: true, latency: 10 * time.Millisecond})
	h.Register(&cacheCheck{name: "redis", hitRate: 0.35})
	h.Register(&apiCheck{name: "billing-api", statusCode: 200, probeJitter: 5 * time.Millisecond})
	h.Register(&diskCheck{name: "data-volume", freeRatio: 0.42})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	status, results := h.Status(ctx)

	fmt.Printf("aggregated status: %s\n\n", status)

	for _, r := range results {
		fmt.Printf("  - %-12s %-9s %s (%.2fms)\n", r.Name, r.Status, r.Message, float64(r.Duration.Microseconds())/1000)
	}
}
