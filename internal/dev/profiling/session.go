// Package profiling captures reviewable performance artifacts for native runs.
package profiling

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"

	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// Session owns one CPU capture and its derived artifacts. Callers must stop the session to release the profile file.
type Session struct {
	outputDirectory     string
	cpuProfile          *os.File
	executable          string
	renderPDF           bool
	mu                  sync.Mutex
	allScenes           bool
	selectedScenes      map[string]bool
	sceneHeapProfiles   map[string][]string
	diagnosticsProvider func() any
}

// Start begins CPU profiling in directory. A successful call transfers ownership of the profile file to the session;
// when renderPDF is true, Stop also renders CPU and live-heap graph views with go tool pprof.
func Start(directory string, renderPDF bool) (*Session, error) {
	expandedDirectory, err := darkpaths.ExpandHost(directory)
	if err != nil {
		return nil, fmt.Errorf("profiling: expand output directory: %w", err)
	}

	if err := os.MkdirAll(expandedDirectory, 0o755); err != nil {
		return nil, fmt.Errorf("profiling: create output directory: %w", err)
	}

	cpuProfile, err := os.Create(filepath.Join(expandedDirectory, "cpu.pprof"))
	if err != nil {
		return nil, fmt.Errorf("profiling: create CPU profile: %w", err)
	}

	if err := pprof.StartCPUProfile(cpuProfile); err != nil {
		_ = cpuProfile.Close()

		return nil, fmt.Errorf("profiling: start CPU profile: %w", err)
	}

	executable, err := os.Executable()
	if err != nil {
		// Starting the process-global profiler commits this function to unwind it on every later failure.
		pprof.StopCPUProfile()

		_ = cpuProfile.Close()

		return nil, fmt.Errorf("profiling: locate executable: %w", err)
	}

	return &Session{
		outputDirectory:   expandedDirectory,
		cpuProfile:        cpuProfile,
		executable:        executable,
		renderPDF:         renderPDF,
		selectedScenes:    make(map[string]bool),
		sceneHeapProfiles: make(map[string][]string),
	}, nil
}

// SetDiagnostics installs a snapshot provider written beside every scene heap
// and once more at shutdown. The provider is called synchronously without the session lock, so it may inspect state
// that reports into the profiler without deadlocking.
func (s *Session) SetDiagnostics(provider func() any) {
	s.mu.Lock()
	s.diagnosticsProvider = provider
	s.mu.Unlock()
}

// ConfigureScenes adds a comma-separated list of scene IDs, or "all", to the capture set. Repeated calls accumulate
// selections so independently configured callers cannot disable an earlier request.
func (s *Session) ConfigureScenes(value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, name := range strings.Split(value, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		if name == "all" {
			s.allScenes = true
		} else {
			s.selectedScenes[name] = true
		}
	}
}

// Stop completes capture and renders cpu.pdf and heap.pdf when requested. Raw
// profiles are preserved even if diagnostics, Graphviz, or PDF rendering fails; all failures are returned together.
func (s *Session) Stop() error {
	if s == nil || s.cpuProfile == nil {
		return nil
	}

	// CPU profiling is process-global, so it must stop before its destination file is closed.
	pprof.StopCPUProfile()

	cpuCloseErr := s.cpuProfile.Close()
	s.cpuProfile = nil

	runtime.GC()

	heapPath := filepath.Join(s.outputDirectory, "heap.pprof")
	heapErr := writeHeapProfile(heapPath)
	diagnosticsErr := s.writeDiagnosticsSnapshot(filepath.Join(s.outputDirectory, "diagnostics.json"))

	if !s.renderPDF {
		return errors.Join(cpuCloseErr, heapErr, diagnosticsErr)
	}

	cpuRenderErr := renderProfilePDF(
		s.executable,
		filepath.Join(s.outputDirectory, "cpu.pprof"),
		filepath.Join(s.outputDirectory, "cpu.pdf"),
		"cpu",
	)
	heapRenderErr := renderProfilePDF(
		s.executable,
		heapPath,
		filepath.Join(s.outputDirectory, "heap.pdf"),
		"inuse_space",
	)
	err := errors.Join(cpuCloseErr, heapErr, diagnosticsErr, cpuRenderErr, heapRenderErr)

	return s.renderSceneProfiles(err)
}
