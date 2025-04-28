//go:build !profile

package profilingutil

import (
	"log"
)

// no-op, except the logging
func StartProfilers(cfg Config) error {
	log.Printf("%s: Profiling disabled: Skipping profiler setup.", packageName)
	return nil
}

// same
func StopProfilers() {
	log.Printf("%s: Profiling disabled: Skipping profiler teardown.", packageName)
}
