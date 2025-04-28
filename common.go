package profilingutil

import (
	"fmt"
	"io"
	"log"
	"os"
)

const packageName = "profilingutil"

// If a writer is nil, that specific profile will not be collected.
// The caller is responsible for managing the lifecycle of these writers (e.g., closing files).
type Config struct {
	CPUProfileWriter       io.Writer
	HeapProfileWriter      io.Writer
	GoroutineProfileWriter io.Writer
}

// CloseAll closes any writers in the Config that implement io.Closer.
// Useful when the Config was created using NewConfigFromPaths.
func (c Config) CloseAll() {
	log.Printf("%s: Closing profiling writers.", packageName)
	var closeErrors []error

	closeWriter := func(w io.Writer, name string) {
		if w != nil {
			if closer, ok := w.(io.Closer); ok {
				if err := closer.Close(); err != nil {
					closeErrors = append(closeErrors, fmt.Errorf("failed to close %s writer: %w", name, err))
				} else {
					log.Printf("%s: Closed %s writer.", packageName, name)
				}
			}
		}
	}

	closeWriter(c.CPUProfileWriter, "CPU profile")
	closeWriter(c.HeapProfileWriter, "Heap profile")
	closeWriter(c.GoroutineProfileWriter, "Goroutine profile")

	// TODO: Errorgroup
	if len(closeErrors) > 0 {
		for _, err := range closeErrors {
			log.Printf("%s: Error during writer close: %v", packageName, err)
		}
	}
}

// NewConfigFromPaths creates a Config by opening files at the given paths.
// If a path is empty, the corresponding writer in the returned Config will be nil.
// Use CloseAll to conveniently close, without referencing io.Writer
func NewConfigFromPaths(cpuPath, heapPath, goroutinePath string) (Config, error) {
	cfg := Config{}
	var err error

	createAndAssign := func(path string, writerField *io.Writer, name string) error {
		if path != "" {
			file, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("%s: failed to create %s file %s: %w", packageName, name, path, err)
			}
			*writerField = file
		}
		return nil
	}

	// TODO: Fix this logic to be less order reliant

	if err = createAndAssign(cpuPath, &cfg.CPUProfileWriter, "CPU profile"); err != nil {
		return Config{}, err
	}

	if err = createAndAssign(heapPath, &cfg.HeapProfileWriter, "Heap profile"); err != nil {
		// If Heap file creation fails, close the CPU file if it was opened
		if cfg.CPUProfileWriter != nil {
			cfg.CPUProfileWriter.(io.Closer).Close()
		}
		return Config{}, err
	}

	if err = createAndAssign(goroutinePath, &cfg.GoroutineProfileWriter, "Goroutine profile"); err != nil {
		// If Goroutine file creation fails, close previously opened files
		if cfg.CPUProfileWriter != nil {
			cfg.CPUProfileWriter.(io.Closer).Close()
		}
		if cfg.HeapProfileWriter != nil {
			cfg.HeapProfileWriter.(io.Closer).Close()
		}
		return Config{}, err
	}

	return cfg, nil
}
