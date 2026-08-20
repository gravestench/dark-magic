package profiling

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
)

// CaptureSceneHeap records live heap state immediately before scene resources are released. Only successful captures
// advance the monotonically numbered artifact sequence, so a failed attempt can be retried at the same path.
func (s *Session) CaptureSceneHeap(name string) error {
	wanted, index := s.sceneHeapRequest(name)
	if !wanted {
		return nil
	}

	directory := filepath.Join(s.outputDirectory, "scenes", safeName(name))
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("profiling: create scene directory: %w", err)
	}

	// A collection immediately before the snapshot keeps these artifacts focused on live ownership.
	runtime.GC()

	heapPath := filepath.Join(directory, fmt.Sprintf("heap-%03d.pprof", index))
	if err := writeHeapProfile(heapPath); err != nil {
		return fmt.Errorf("profiling: capture scene %q heap: %w", name, err)
	}

	diagnosticsPath := filepath.Join(directory, fmt.Sprintf("diagnostics-%03d.json", index))
	if err := s.writeDiagnosticsSnapshot(diagnosticsPath); err != nil {
		return err
	}

	s.recordSceneHeapProfile(name, heapPath)

	return nil
}

// sceneHeapRequest snapshots capture policy and the next successful index without holding the lock across GC or I/O.
// Callers must serialize captures for the same scene because an index is not reserved until its artifacts succeed.
func (s *Session) sceneHeapRequest(name string) (bool, int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	wanted := s.allScenes || s.selectedScenes[name]
	index := len(s.sceneHeapProfiles[name]) + 1

	return wanted, index
}

// recordSceneHeapProfile publishes a completed heap path only after its paired diagnostics write succeeds. That keeps
// rendering and the next capture index aligned with complete scene visits.
func (s *Session) recordSceneHeapProfile(name, path string) {
	s.mu.Lock()
	s.sceneHeapProfiles[name] = append(s.sceneHeapProfiles[name], path)
	s.mu.Unlock()
}

// writeDiagnosticsSnapshot serializes one provider result with stable indentation and a trailing newline. The provider
// runs outside the session lock so diagnostics may safely inspect components that also use the profiling session.
func (s *Session) writeDiagnosticsSnapshot(path string) error {
	s.mu.Lock()
	provider := s.diagnosticsProvider
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

// writeHeapProfile owns the destination file through both the heap write and close. Joining both failures preserves a
// close error without hiding an earlier profile error.
func writeHeapProfile(path string) error {
	profile, err := os.Create(path)
	if err != nil {
		return err
	}

	writeErr := pprof.WriteHeapProfile(profile)

	return errors.Join(writeErr, profile.Close())
}

// safeName replaces every non-ASCII filename character with one underscore. This one-rune-to-one-rune mapping keeps
// scene artifact paths deterministic while preventing separators and traversal syntax from escaping their directory.
func safeName(name string) string {
	return strings.Map(safeSceneNameRune, name)
}

// safeSceneNameRune admits only the portable filename alphabet used by scene IDs; every other rune maps predictably.
func safeSceneNameRune(character rune) rune {
	if character >= 'a' && character <= 'z' ||
		character >= 'A' && character <= 'Z' ||
		character >= '0' && character <= '9' ||
		character == '-' || character == '_' {
		return character
	}

	return '_'
}
