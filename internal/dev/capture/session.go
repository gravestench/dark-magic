// Package capture records local visual-review artifacts without embedding game
// imagery in tests or source control.
package capture

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	defaultDirectory = "./captures/frontend"
	defaultScenes    = "loading,title"
	reportFilename   = "report.json"
	reportVersion    = 1
)

// Defaults makes either capture flag sufficient to opt into capture mode while preserving explicitly supplied values.
func Defaults(directory, scenes string) (string, string) {
	directory = strings.TrimSpace(directory)
	scenes = strings.TrimSpace(scenes)

	switch {
	case directory == "" && scenes != "":
		directory = defaultDirectory
	case directory != "" && scenes == "":
		scenes = defaultScenes
	}

	return directory, scenes
}

// Screenshotter writes the current framebuffer to the requested path, leaving file ownership with the capture session.
type Screenshotter interface {
	CaptureScreenshot(string) error
}

// Session captures the first stable frame for each requested scene.
type Session struct {
	directory          string
	requestedScenes    map[string]bool
	capturedScenes     map[string]bool
	settleFrames       int
	currentScene       string
	stableFrames       int
	structuralRevision uint64
	results            []Result
	screenshotter      Screenshotter
	captureErr         error
}

// New validates capture policy before creating its output directory, so invalid sessions have no filesystem effects.
func New(directory, scenes string, settleFrames int, screenshotter Screenshotter) (*Session, error) {
	if err := validateSessionConfiguration(directory, settleFrames, screenshotter); err != nil {
		return nil, err
	}

	requestedScenes := parseRequestedScenes(scenes)
	if len(requestedScenes) == 0 {
		return nil, fmt.Errorf("capture: at least one scene is required")
	}

	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, err
	}

	return &Session{
		directory:       directory,
		requestedScenes: requestedScenes,
		capturedScenes:  make(map[string]bool),
		settleFrames:    settleFrames,
		screenshotter:   screenshotter,
	}, nil
}

// validateSessionConfiguration preserves constructor error precedence before scene parsing or directory creation.
func validateSessionConfiguration(directory string, settleFrames int, screenshotter Screenshotter) error {
	if directory == "" || screenshotter == nil {
		return fmt.Errorf("capture: directory and screenshotter are required")
	}

	if settleFrames < 1 {
		return fmt.Errorf("capture: settle frames must be positive")
	}

	return nil
}

// parseRequestedScenes trims and deduplicates names so completion counts each logical scene only once.
func parseRequestedScenes(scenes string) map[string]bool {
	requestedScenes := make(map[string]bool)

	for _, scene := range strings.Split(scenes, ",") {
		scene = strings.TrimSpace(scene)
		if scene != "" {
			requestedScenes[scene] = true
		}
	}

	return requestedScenes
}

// Observe advances scene stability after a rendered frame and captures only a complete, idle presentation.
func (session *Session) Observe(stack []string, structuralRevision uint64, busy bool) {
	if session.captureErr != nil || len(stack) == 0 {
		return
	}

	scene := stack[len(stack)-1]
	if scene != session.currentScene {
		if err := session.beginSceneObservation(scene, structuralRevision); err != nil {
			session.captureErr = err
			return
		}
	}

	if structuralRevision != session.structuralRevision {
		// Retained-node changes can arrive after navigation. Restarting the window prevents a partial scene from being
		// certified, while texture and text updates intentionally leave the structural revision unchanged.
		session.structuralRevision = structuralRevision
		session.stableFrames = 0
	}

	if busy {
		// Worker residency is part of scene assembly even before it mutates retained topology. Resetting here prevents a
		// slow worker from allowing the settle window to expire against a partial frame.
		session.stableFrames = 0
		return
	}

	session.stableFrames++
	if !session.sceneReadyForCapture(scene) {
		return
	}

	session.captureScene(scene)
}

// beginSceneObservation rejects requested scenes that disappear before capture and resets state for valid transitions.
func (session *Session) beginSceneObservation(scene string, structuralRevision uint64) error {
	if session.requestedScenes[session.currentScene] && !session.capturedScenes[session.currentScene] {
		return fmt.Errorf(
			"scene %q transitioned before a visible frame could be captured",
			session.currentScene,
		)
	}

	session.currentScene = scene
	session.stableFrames = 0
	session.structuralRevision = structuralRevision

	return nil
}

// sceneReadyForCapture centralizes the gate that prevents duplicate, unrequested, or unsettled screenshots.
func (session *Session) sceneReadyForCapture(scene string) bool {
	return session.requestedScenes[scene] &&
		!session.capturedScenes[scene] &&
		session.stableFrames >= session.settleFrames
}

// captureScene adds one artifact to session state only after its decoded pixels pass visibility checks.
func (session *Session) captureScene(scene string) {
	name := screenshotFilename(len(session.results)+1, scene)
	path := filepath.Join(session.directory, name)

	if err := session.screenshotter.CaptureScreenshot(path); err != nil {
		session.captureErr = err
		return
	}

	result, err := inspectScreenshot(path, scene, name)
	if errors.Is(err, errBlankFrame) {
		// A blank framebuffer is transient rather than terminal. Removal is best-effort so the next observation retries
		// the same sequence number and path without replacing the useful inspection result with a cleanup error.
		_ = os.Remove(path)
		return
	}

	if err != nil {
		session.captureErr = err
		return
	}

	session.capturedScenes[scene] = true
	session.results = append(session.results, result)
}

// Complete reports success only when no terminal capture error occurred and every requested scene has an artifact.
func (session *Session) Complete() bool {
	return session.captureErr == nil && len(session.capturedScenes) == len(session.requestedScenes)
}

// Close writes the report even for failed sessions, then preserves capture and incompleteness errors over report
// errors.
func (session *Session) Close() error {
	reportErr := session.writeReport()

	if session.captureErr != nil {
		return fmt.Errorf("capture: %w", session.captureErr)
	}

	missingScenes := session.missingScenes()
	if len(missingScenes) != 0 {
		return fmt.Errorf("capture: incomplete; missing scenes: %s", strings.Join(missingScenes, ","))
	}

	return reportErr
}

// writeReport snapshots result order and retains the indented, newline-terminated on-disk format expected by tools.
func (session *Session) writeReport() error {
	report := Report{
		Version: reportVersion,
		Results: append([]Result(nil), session.results...),
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')

	return os.WriteFile(filepath.Join(session.directory, reportFilename), data, 0o644)
}

// missingScenes returns a sorted diagnostic list so map iteration cannot make Close errors nondeterministic.
func (session *Session) missingScenes() []string {
	missingScenes := make([]string, 0, len(session.requestedScenes)-len(session.capturedScenes))

	for scene := range session.requestedScenes {
		if !session.capturedScenes[scene] {
			missingScenes = append(missingScenes, scene)
		}
	}

	slices.Sort(missingScenes)

	return missingScenes
}
