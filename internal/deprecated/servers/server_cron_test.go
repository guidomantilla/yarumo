package servers

import (
	"context"
	"testing"
	"time"
)

type fakeCron struct {
	started bool
	stopped bool
}

func (f *fakeCron) Start() { f.started = true }

func (f *fakeCron) Stop() { f.stopped = true }

func TestBuildCronServer(t *testing.T) {
	fc := &fakeCron{}
	name, srv := BuildCronServer(fc)
	if name != "cron-server" || srv == nil {
		t.Fatalf("unexpected: name=%s srv=%v", name, srv)
	}
}

func TestNewCronServer_AssertionsCoverage(t *testing.T) {
	// Exercise assert for nil cron (it only logs)
	_ = NewCronServer(nil)
}

func TestCronServer_RunAndStop(t *testing.T) {
	fc := &fakeCron{}
	cs := NewCronServer(fc).(*cronServer)

	ctx := context.Background()
	done := make(chan error, 1)
	go func() { done <- cs.Run(ctx) }()

	time.Sleep(50 * time.Millisecond)
	if !fc.started {
		t.Fatal("expected Start to be called")
	}

	if err := cs.Stop(ctx); err != nil {
		t.Fatalf("stop returned error: %v", err)
	}
	if !fc.stopped {
		t.Fatal("expected Stop to be called")
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run returned error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("run did not finish after stop")
	}
}
