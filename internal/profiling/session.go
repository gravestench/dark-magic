// Package profiling captures reviewable performance artifacts for native runs.
package profiling

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"sync"

	darkpaths "github.com/gravestench/dark-magic/pkg/paths"
)

// Session owns one CPU capture and writes a heap snapshot when stopped.
type Session struct {
	directory   string
	cpu         *os.File
	binary      string
	renderPDF   bool
	mu          sync.Mutex
	allScenes   bool
	scenes      map[string]bool
	heaps       map[string][]string
	diagnostics func() any
}

// Start begins CPU profiling in directory. When renderPDF is true, Stop also
// renders CPU and live-heap graph views with go tool pprof.
func Start(directory string, renderPDF bool) (*Session, error) {
	expanded, err := darkpaths.ExpandHost(directory)
	if err != nil {
		return nil, fmt.Errorf("profiling: expand output directory: %w", err)
	}
	if err := os.MkdirAll(expanded, 0o755); err != nil {
		return nil, fmt.Errorf("profiling: create output directory: %w", err)
	}
	cpu, err := os.Create(filepath.Join(expanded, "cpu.pprof"))
	if err != nil {
		return nil, fmt.Errorf("profiling: create CPU profile: %w", err)
	}
	if err := pprof.StartCPUProfile(cpu); err != nil {
		_ = cpu.Close()
		return nil, fmt.Errorf("profiling: start CPU profile: %w", err)
	}
	binary, err := os.Executable()
	if err != nil {
		pprof.StopCPUProfile()
		_ = cpu.Close()
		return nil, fmt.Errorf("profiling: locate executable: %w", err)
	}
	return &Session{directory: expanded, cpu: cpu, binary: binary, renderPDF: renderPDF, scenes: make(map[string]bool), heaps: make(map[string][]string)}, nil
}

// SetDiagnostics installs a snapshot provider written beside every scene heap
// and once more at shutdown. The provider must be safe to call synchronously.
func (s *Session) SetDiagnostics(provider func() any) {
	s.mu.Lock()
	s.diagnostics = provider
	s.mu.Unlock()
}

// ConfigureScenes selects a comma-separated list of scene IDs, or "all".
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
			s.scenes[name] = true
		}
	}
}

// CaptureSceneHeap records live heap state immediately before scene resources
// are released. Repeated visits receive monotonically numbered snapshots.
func (s *Session) CaptureSceneHeap(name string) error {
	s.mu.Lock()
	wanted := s.allScenes || s.scenes[name]
	index := len(s.heaps[name]) + 1
	s.mu.Unlock()
	if !wanted {
		return nil
	}
	directory := filepath.Join(s.directory, "scenes", safeName(name))
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("profiling: create scene directory: %w", err)
	}
	runtime.GC()
	path := filepath.Join(directory, fmt.Sprintf("heap-%03d.pprof", index))
	file, err := os.Create(path)
	if err == nil {
		err = pprof.WriteHeapProfile(file)
		err = errors.Join(err, file.Close())
	}
	if err != nil {
		return fmt.Errorf("profiling: capture scene %q heap: %w", name, err)
	}
	if err := s.writeDiagnostics(filepath.Join(directory, fmt.Sprintf("diagnostics-%03d.json", index))); err != nil {
		return err
	}
	s.mu.Lock()
	s.heaps[name] = append(s.heaps[name], path)
	s.mu.Unlock()
	return nil
}

// Stop completes capture and renders cpu.pdf and heap.pdf when requested. Raw
// profiles are preserved even if Graphviz or PDF rendering is unavailable.
func (s *Session) Stop() error {
	if s == nil || s.cpu == nil {
		return nil
	}
	pprof.StopCPUProfile()
	err := s.cpu.Close()
	s.cpu = nil
	runtime.GC()
	heapPath := filepath.Join(s.directory, "heap.pprof")
	heap, heapErr := os.Create(heapPath)
	if heapErr == nil {
		heapErr = pprof.WriteHeapProfile(heap)
		heapErr = errors.Join(heapErr, heap.Close())
	}
	if !s.renderPDF {
		return errors.Join(err, heapErr, s.writeDiagnostics(filepath.Join(s.directory, "diagnostics.json")))
	}
	err = errors.Join(err, heapErr, s.writeDiagnostics(filepath.Join(s.directory, "diagnostics.json")),
		render(s.binary, filepath.Join(s.directory, "cpu.pprof"), filepath.Join(s.directory, "cpu.pdf"), "cpu"),
		render(s.binary, heapPath, filepath.Join(s.directory, "heap.pdf"), "inuse_space"))
	s.mu.Lock()
	names := make([]string, 0, len(s.heaps))
	for name := range s.heaps {
		names = append(names, name)
	}
	sort.Strings(names)
	heaps := make(map[string][]string, len(s.heaps))
	for _, name := range names {
		heaps[name] = append([]string(nil), s.heaps[name]...)
	}
	s.mu.Unlock()
	for _, name := range names {
		directory := filepath.Join(s.directory, "scenes", safeName(name))
		err = errors.Join(err, render(s.binary, filepath.Join(s.directory, "cpu.pprof"), filepath.Join(directory, "cpu.pdf"), "cpu", "-tagfocus=scene="+regexp.QuoteMeta(name)))
		for index, profile := range heaps[name] {
			err = errors.Join(err, render(s.binary, profile, filepath.Join(directory, fmt.Sprintf("heap-%03d.pdf", index+1)), "inuse_space"))
		}
	}
	return err
}

func (s *Session) writeDiagnostics(path string) error {
	s.mu.Lock()
	provider := s.diagnostics
	s.mu.Unlock()
	if provider == nil {
		return nil
	}
	data, err := json.MarshalIndent(provider(), "", "  ")
	if err != nil {
		return fmt.Errorf("profiling: encode diagnostics: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("profiling: write diagnostics: %w", err)
	}
	return nil
}

func safeName(name string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)
}

func render(binary, profile, output, sampleIndex string, filters ...string) error {
	arguments := []string{"tool", "pprof", "-pdf", "-sample_index=" + sampleIndex, "-output", output}
	arguments = append(arguments, filters...)
	arguments = append(arguments, binary, profile)
	command := exec.Command("go", arguments...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("profiling: render %s: %w: %s", filepath.Base(output), err, stderr.String())
	}
	return nil
}
