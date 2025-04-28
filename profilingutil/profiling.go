//go:build profile

package profilingutil

import (
	"fmt"
	"io"
	"log"
	"runtime"
	"runtime/pprof"
)

var (
	cpuWriter       io.Writer
	heapWriter      io.Writer
	goroutineWriter io.Writer
)

func StartProfilers(cfg Config) error {
	log.Printf("%s: Profiling enabled: Setting up profilers.", packageName)

	cpuWriter = cfg.CPUProfileWriter
	heapWriter = cfg.HeapProfileWriter
	goroutineWriter = cfg.GoroutineProfileWriter

	if cpuWriter != nil {
		if err := pprof.StartCPUProfile(cpuWriter); err != nil {
			log.Printf("%s: could not start CPU profile: %v", packageName, err)
			cpuWriter = nil // Indicate profiling didn't start
			return fmt.Errorf("%s: failed to start CPU profile: %w", packageName, err)
		} else {
			log.Printf("%s: CPU profiling started.", packageName)
		}
	} else {
		log.Printf("%s: CPUProfileWriter not provided, skipping CPU profiling.", packageName)
	}

	if heapWriter != nil || goroutineWriter != nil {
		log.Printf("%s: Heap and Goroutine profiles will be written on exit.", packageName)
	}

	return nil // Return nil if setup was successful or no CPU writer was provided
}

// The caller is responsible for closing the writers if they are io.Closer.
func StopProfilers() {
	log.Printf("%s: Profiling enabled: Tearing down profilers.", packageName)

	if cpuWriter != nil {
		pprof.StopCPUProfile()
		log.Printf("%s: CPU profile written.", packageName)
		// Do NOT close cpuWriter here. The caller owns it.
		cpuWriter = nil // Reset after use
	}

	if heapWriter != nil {
		runtime.GC() // Get up-to-date statistics
		if err := pprof.WriteHeapProfile(heapWriter); err != nil {
			log.Printf("%s: could not write heap profile: %v", packageName, err)
		} else {
			log.Printf("%s: Heap profile written.", packageName)
		}
		// Do NOT close heapWriter here. The caller owns it.
		heapWriter = nil // Reset after use
	}

	if goroutineWriter != nil {
		if err := pprof.Lookup("goroutine").WriteTo(goroutineWriter, 0); err != nil {
			log.Printf("%s: could not write goroutine profile: %v", packageName, err)
		} else {
			log.Printf("%s: Goroutine profile written.", packageName)
		}
		// Do NOT close goroutineWriter here. The caller owns it.
		goroutineWriter = nil // Reset after use
	}
}
