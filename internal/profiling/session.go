// Package profiling captures reviewable performance artifacts for native runs.
package profiling

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/pprof"

	darkpaths "github.com/gravestench/dark-magic/pkg/paths"
)

// Session owns one CPU capture and writes a heap snapshot when stopped.
type Session struct {
	directory string
	cpu       *os.File
	binary    string
	renderPDF bool
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
	return &Session{directory: expanded, cpu: cpu, binary: binary, renderPDF: renderPDF}, nil
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
		return errors.Join(err, heapErr)
	}
	err = errors.Join(err, heapErr,
		render(s.binary, filepath.Join(s.directory, "cpu.pprof"), filepath.Join(s.directory, "cpu.pdf"), "cpu"),
		render(s.binary, heapPath, filepath.Join(s.directory, "heap.pdf"), "inuse_space"))
	return err
}

func render(binary, profile, output, sampleIndex string) error {
	command := exec.Command("go", "tool", "pprof", "-pdf", "-sample_index="+sampleIndex, "-output", output, binary, profile)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("profiling: render %s: %w: %s", filepath.Base(output), err, stderr.String())
	}
	return nil
}
