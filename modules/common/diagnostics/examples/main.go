// Example demonstrates each pprof capture wrapper in the diagnostics package.
//
// Run from the module root:
//
//	cd modules/common && go run ./diagnostics/examples
//
// Each profile is captured to a file under the current working directory.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/guidomantilla/yarumo/common/diagnostics"
)

func main() {
	heapExample()
	goroutineExample()
	blockExample()
	cpuExample()
}

// heapExample captures the current heap profile snapshot. The heap profile is
// sampled by the Go runtime by default, so this call has near-zero overhead.
func heapExample() {
	fmt.Println("=== Heap Profile ===")

	// Force one allocation so the heap profile is non-empty for the demo.
	sink := make([]byte, 1<<20)
	runtime.KeepAlive(sink)
	runtime.GC()

	file, err := os.CreateTemp("", "heap-*.pprof")
	if err != nil {
		log.Fatalf("create temp: %v", err)
	}

	defer func() {
		_ = file.Close()
	}()

	err = diagnostics.CaptureHeapProfile(file)
	if err != nil {
		log.Fatalf("CaptureHeapProfile: %v", err)
	}

	fmt.Printf("heap profile written to %s\n\n", file.Name())
}

// goroutineExample captures the goroutine profile snapshot. Cost scales with
// the number of live goroutines.
func goroutineExample() {
	fmt.Println("=== Goroutine Profile ===")

	file, err := os.CreateTemp("", "goroutine-*.pprof")
	if err != nil {
		log.Fatalf("create temp: %v", err)
	}

	defer func() {
		_ = file.Close()
	}()

	err = diagnostics.CaptureGoroutineProfile(file)
	if err != nil {
		log.Fatalf("CaptureGoroutineProfile: %v", err)
	}

	fmt.Printf("goroutine profile written to %s\n\n", file.Name())
}

// blockExample captures the block profile. Block profiling MUST be enabled
// beforehand via runtime.SetBlockProfileRate — without it the profile is
// empty. Block profiling carries non-trivial overhead (every blocking event
// is sampled), so production code typically enables it for short windows.
func blockExample() {
	fmt.Println("=== Block Profile ===")

	// Sample every blocking event for the demo.
	runtime.SetBlockProfileRate(1)
	defer runtime.SetBlockProfileRate(0)

	// Provoke a blocking event so the profile contains samples.
	ch := make(chan struct{})

	go func() {
		time.Sleep(5 * time.Millisecond)
		close(ch)
	}()
	<-ch

	file, err := os.CreateTemp("", "block-*.pprof")
	if err != nil {
		log.Fatalf("create temp: %v", err)
	}

	defer func() {
		_ = file.Close()
	}()

	err = diagnostics.CaptureBlockProfile(file)
	if err != nil {
		log.Fatalf("CaptureBlockProfile: %v", err)
	}

	fmt.Printf("block profile written to %s\n\n", file.Name())
}

// cpuExample captures a CPU profile for a short window. CPU profiling has
// low overhead (~5%) and is the canonical "always available" profile.
func cpuExample() {
	fmt.Println("=== CPU Profile ===")

	file, err := os.CreateTemp("", "cpu-*.pprof")
	if err != nil {
		log.Fatalf("create temp: %v", err)
	}

	defer func() {
		_ = file.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err = diagnostics.CaptureCPUProfile(ctx, file, 200*time.Millisecond)
	if err != nil {
		log.Fatalf("CaptureCPUProfile: %v", err)
	}

	fmt.Printf("cpu profile written to %s\n\n", file.Name())
}
