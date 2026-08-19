package profiling

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
)

// renderSceneProfiles renders the completed scene captures in deterministic name and visit order. It carries earlier
// shutdown errors forward so optional rendering failures cannot hide raw-profile or diagnostics failures.
func (s *Session) renderSceneProfiles(err error) error {
	names, heapProfiles := s.snapshotSceneHeapProfiles()
	for _, name := range names {
		directory := filepath.Join(s.outputDirectory, "scenes", safeName(name))
		sceneFilter := "-tagfocus=scene=" + regexp.QuoteMeta(name)
		err = errors.Join(err, renderProfilePDF(
			s.executable,
			filepath.Join(s.outputDirectory, "cpu.pprof"),
			filepath.Join(directory, "cpu.pdf"),
			"cpu",
			sceneFilter,
		))

		for index, profile := range heapProfiles[name] {
			output := filepath.Join(directory, fmt.Sprintf("heap-%03d.pdf", index+1))
			err = errors.Join(err, renderProfilePDF(s.executable, profile, output, "inuse_space"))
		}
	}

	return err
}

// snapshotSceneHeapProfiles returns sorted names and defensive copies of their paths. Rendering can then run without
// holding the session lock or observing slices that a later completed capture mutates.
func (s *Session) snapshotSceneHeapProfiles() ([]string, map[string][]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	names := make([]string, 0, len(s.sceneHeapProfiles))
	for name := range s.sceneHeapProfiles {
		names = append(names, name)
	}

	sort.Strings(names)

	heapProfiles := make(map[string][]string, len(s.sceneHeapProfiles))
	for _, name := range names {
		heapProfiles[name] = append([]string(nil), s.sceneHeapProfiles[name]...)
	}

	return names, heapProfiles
}

// renderProfilePDF invokes pprof with filters before its positional binary and profile arguments. Stderr is included
// verbatim in failures because missing Graphviz and invalid profiles otherwise produce indistinguishable exit errors.
func renderProfilePDF(executable, profile, output, sampleIndex string, filters ...string) error {
	arguments := []string{"tool", "pprof", "-pdf", "-sample_index=" + sampleIndex, "-output", output}
	arguments = append(arguments, filters...)
	arguments = append(arguments, executable, profile)

	command := exec.Command("go", arguments...)

	var stderr bytes.Buffer

	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("profiling: render %s: %w: %s", filepath.Base(output), err, stderr.String())
	}

	return nil
}
