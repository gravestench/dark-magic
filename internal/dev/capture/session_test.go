package capture

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeScreenshotter emits deterministic PNGs and can model transient blank framebuffers without external rendering.
type fakeScreenshotter struct {
	calls      int
	blankCalls int
}

// CaptureScreenshot transfers ownership of a closed PNG to the session so inspection observes complete file contents.
func (screenshotter *fakeScreenshotter) CaptureScreenshot(name string) error {
	screenshotter.calls++

	file, err := os.Create(name)
	if err != nil {
		return err
	}

	canvas := image.NewRGBA(image.Rect(0, 0, 8, 6))
	if screenshotter.calls > screenshotter.blankCalls {
		canvas.Set(0, 0, color.White)
	}

	// Closing even after an encode failure prevents fixture descriptors from leaking; the encoding error remains primary.
	encodeErr := png.Encode(file, canvas)
	closeErr := file.Close()

	if encodeErr != nil {
		return encodeErr
	}

	return closeErr
}

// TestDefaults verifies either non-empty flag opts into capture without replacing explicitly supplied policy.
func TestDefaults(t *testing.T) {
	tests := []struct {
		name          string
		directory     string
		scenes        string
		wantDirectory string
		wantScenes    string
	}{
		{name: "disabled"},
		{
			name:          "directory enables default scenes",
			directory:     "./captures/custom",
			wantDirectory: "./captures/custom",
			wantScenes:    "loading,title",
		},
		{
			name:          "scenes enable default directory",
			scenes:        "death",
			wantDirectory: "./captures/frontend",
			wantScenes:    "death",
		},
		{
			name:          "explicit",
			directory:     "./captures/custom",
			scenes:        "death",
			wantDirectory: "./captures/custom",
			wantScenes:    "death",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory, scenes := Defaults(test.directory, test.scenes)
			if directory != test.wantDirectory || scenes != test.wantScenes {
				t.Fatalf(
					"Defaults(%q, %q) = %q, %q; want %q, %q",
					test.directory,
					test.scenes,
					directory,
					scenes,
					test.wantDirectory,
					test.wantScenes,
				)
			}
		})
	}
}

// TestSessionRetriesBlankFramebuffer verifies a transient blank frame neither completes nor poisons the session.
func TestSessionRetriesBlankFramebuffer(t *testing.T) {
	directory := t.TempDir()
	screenshotter := &fakeScreenshotter{blankCalls: 1}
	session := newTestSession(t, directory, "loading", 1, screenshotter)

	session.Observe([]string{"loading"}, 1, false)

	if len(session.results) != 0 {
		t.Fatalf("blank frame was captured: %#v", session.results)
	}

	if session.Complete() {
		t.Fatal("session completed after a rejected blank frame")
	}

	// The same stable scene must retry immediately because blank pixels are transient, not a new settling event.
	session.Observe([]string{"loading"}, 1, false)

	if err := session.Close(); err != nil {
		t.Fatal(err)
	}

	if screenshotter.calls != 2 || len(session.results) != 1 {
		t.Fatalf("calls = %d; results = %#v", screenshotter.calls, session.results)
	}

	if !session.Complete() {
		t.Fatal("session did not complete after capturing every requested scene")
	}
}

// TestSessionCloseRejectsIncompleteCapture verifies Close deterministically identifies each still-missing scene.
func TestSessionCloseRejectsIncompleteCapture(t *testing.T) {
	session := newTestSession(t, t.TempDir(), "death,title", 1, &fakeScreenshotter{})

	session.Observe([]string{"death"}, 1, false)
	err := session.Close()

	if err == nil || !strings.Contains(err.Error(), "missing scenes: title") {
		t.Fatalf("Close error = %v, want missing title", err)
	}
}

// TestObserveReportsRequestedSceneThatTransitionsBeforeCapture protects the contract against silent artifact loss.
func TestObserveReportsRequestedSceneThatTransitionsBeforeCapture(t *testing.T) {
	session := newTestSession(t, t.TempDir(), "loading", 2, &fakeScreenshotter{})

	session.Observe([]string{"loading"}, 1, false)
	session.Observe([]string{"title"}, 2, false)

	if session.captureErr == nil || !strings.Contains(session.captureErr.Error(), "transitioned before") {
		t.Fatalf("expected transition error, got %v", session.captureErr)
	}
}

// TestSessionCapturesFirstStableRequestedScene verifies artifact order, naming, dimensions, and digest metadata.
func TestSessionCapturesFirstStableRequestedScene(t *testing.T) {
	directory := t.TempDir()
	screenshotter := &fakeScreenshotter{}
	session := newTestSession(t, directory, "loading,title", 2, screenshotter)

	observeSceneFrames(session, "loading", 1, false, 3)
	observeSceneFrames(session, "title", 2, false, 2)

	if err := session.Close(); err != nil {
		t.Fatal(err)
	}

	if screenshotter.calls != 2 {
		t.Fatalf("screenshot calls = %d", screenshotter.calls)
	}

	assertCaptureArtifactsExist(t, directory, "01-loading.png", "02-title.png", "report.json")
	assertFirstResultMetadata(t, session.results)
}

// TestSessionRestartsSettleWindowAfterStructuralChange prevents captures of late retained-node assembly.
func TestSessionRestartsSettleWindowAfterStructuralChange(t *testing.T) {
	screenshotter := &fakeScreenshotter{}
	session := newTestSession(t, t.TempDir(), "vendor", 2, screenshotter)

	session.Observe([]string{"vendor"}, 10, false)
	session.Observe([]string{"vendor"}, 11, false)

	if screenshotter.calls != 0 {
		t.Fatal("captured immediately after a late structural change")
	}

	session.Observe([]string{"vendor"}, 11, false)

	if screenshotter.calls != 1 || !session.Complete() {
		t.Fatalf("stable revised scene was not captured: calls=%d", screenshotter.calls)
	}
}

// TestSessionWaitsForBackgroundSceneAssembly ensures a busy worker repeatedly resets the visual stability window.
func TestSessionWaitsForBackgroundSceneAssembly(t *testing.T) {
	screenshotter := &fakeScreenshotter{}
	session := newTestSession(t, t.TempDir(), "game_world", 2, screenshotter)

	observeSceneFrames(session, "game_world", 10, true, 10)

	if screenshotter.calls != 0 {
		t.Fatal("captured while world residency work was pending")
	}

	observeSceneFrames(session, "game_world", 10, false, 2)

	if screenshotter.calls != 1 || !session.Complete() {
		t.Fatalf("settled world was not captured: calls=%d", screenshotter.calls)
	}
}

// newTestSession keeps constructor failure diagnostics at the fixture boundary and cleanup ownership in each test.
func newTestSession(
	t *testing.T,
	directory string,
	scenes string,
	settleFrames int,
	screenshotter Screenshotter,
) *Session {
	t.Helper()

	session, err := New(directory, scenes, settleFrames, screenshotter)
	if err != nil {
		t.Fatal(err)
	}

	return session
}

// observeSceneFrames expresses repeated post-frame observations without hiding scene, revision, or busy policy.
func observeSceneFrames(
	session *Session,
	scene string,
	structuralRevision uint64,
	busy bool,
	frames int,
) {
	for range frames {
		session.Observe([]string{scene}, structuralRevision, busy)
	}
}

// assertCaptureArtifactsExist preserves per-file failure diagnostics for the complete expected output set.
func assertCaptureArtifactsExist(t *testing.T, directory string, names ...string) {
	t.Helper()

	for _, name := range names {
		if _, err := os.Stat(filepath.Join(directory, name)); err != nil {
			t.Errorf("expected capture artifact %q: %v", name, err)
		}
	}
}

// assertFirstResultMetadata verifies decoded dimensions and digest shape without coupling the test to PNG encoder
// bytes.
func assertFirstResultMetadata(t *testing.T, results []Result) {
	t.Helper()

	if len(results) != 2 ||
		results[0].Width != 8 ||
		results[0].Height != 6 ||
		len(results[0].SHA256) != 64 {
		t.Fatalf("capture results = %#v", results)
	}
}
