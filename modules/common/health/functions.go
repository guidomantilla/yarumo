package health

import (
	"context"
	"time"

	cutils "github.com/guidomantilla/yarumo/common/utils"
)

// aggregate combines a slice of [Result] values into a single [Status] using
// the worst-status-wins rule documented on the package. When results is
// empty, [StatusUnknown] is returned.
func aggregate(results []Result) Status {
	worst := StatusUnknown
	for _, r := range results {
		if r.Status > worst {
			worst = r.Status
		}
	}

	return worst
}

// probeOne runs a single [Check.Probe] with timing instrumentation. Nil
// checks and nil contexts are guarded — callers (the [Health] aggregator)
// must not pass nil. Returned [Result.Duration] is the wall-clock time of
// the probe call, and the result Name is always populated from check.Name().
func probeOne(ctx context.Context, check Check) Result {
	if cutils.Nil(check) {
		return Result{Status: StatusUnknown, Message: ErrCheckNil.Error()}
	}

	if cutils.Nil(ctx) {
		return Result{Name: check.Name(), Status: StatusUnknown, Message: ErrContextNil.Error()}
	}

	start := time.Now()
	res := check.Probe(ctx)
	elapsed := time.Since(start)

	// Force the canonical Name and Duration regardless of what the probe set —
	// the aggregator owns those fields and uses them for downstream reporting.
	res.Name = check.Name()
	res.Duration = elapsed

	return res
}
