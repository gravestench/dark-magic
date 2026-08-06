package capture

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

type fakeScreenshotter struct{ calls int }

func (f *fakeScreenshotter) CaptureScreenshot(name string) error {
	f.calls++
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	canvas := image.NewRGBA(image.Rect(0, 0, 8, 6))
	canvas.Set(0, 0, color.White)
	encodeErr, closeErr := png.Encode(file, canvas), file.Close()
	if encodeErr != nil {
		return encodeErr
	}
	return closeErr
}

func TestSessionCapturesFirstStableRequestedScene(t *testing.T) {
	directory := t.TempDir()
	capturer := &fakeScreenshotter{}
	session, err := New(directory, "loading,title", 2, capturer)
	if err != nil {
		t.Fatal(err)
	}
	session.Observe([]string{"loading"})
	session.Observe([]string{"loading"})
	session.Observe([]string{"loading"})
	session.Observe([]string{"title"})
	session.Observe([]string{"title"})
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
	if capturer.calls != 2 {
		t.Fatalf("screenshot calls = %d", capturer.calls)
	}
	for _, name := range []string{"01-loading.png", "02-title.png", "report.json"} {
		if _, err := os.Stat(filepath.Join(directory, name)); err != nil {
			t.Fatal(err)
		}
	}
	if len(session.results) != 2 || session.results[0].Width != 8 || session.results[0].Height != 6 || len(session.results[0].SHA256) != 64 {
		t.Fatalf("capture results = %#v", session.results)
	}
}
