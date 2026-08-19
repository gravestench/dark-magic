package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestReadLayeredAudioAsset verifies that extraction honors the configured content root rather than bypassing the
// layered filesystem with direct host I/O.
func TestReadLayeredAudioAsset(t *testing.T) {
	contentRoot := t.TempDir()
	assetPath := "data/global/sfx/readability.wav"

	hostPath := filepath.Join(contentRoot, filepath.FromSlash(assetPath))
	if err := os.MkdirAll(filepath.Dir(hostPath), 0o755); err != nil {
		t.Fatal(err)
	}

	want := []byte("test-wave-data")
	if err := os.WriteFile(hostPath, want, 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MPQ_DIRECTORY", contentRoot)

	got, err := readLayeredAudioAsset(assetPath)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("readLayeredAudioAsset(%q) = %q, want %q", assetPath, got, want)
	}
}

// TestResolveAudioDestinationWithoutOutput verifies that inspection-only runs do not allocate an output file.
func TestResolveAudioDestinationWithoutOutput(t *testing.T) {
	destination, err := resolveAudioDestination("", false)
	if err != nil {
		t.Fatal(err)
	}

	if destination != (audioDestination{}) {
		t.Fatalf("resolveAudioDestination(\"\", false) = %+v, want no destination", destination)
	}
}

// TestResolveAudioDestinationExpandsRequestedPath verifies that user-facing aliases are resolved before any write and
// that explicit output remains caller-owned.
func TestResolveAudioDestinationExpandsRequestedPath(t *testing.T) {
	outputRoot := t.TempDir()
	t.Setenv("DARK_MAGIC_AUDIO_OUTPUT", outputRoot)

	destination, err := resolveAudioDestination("$DARK_MAGIC_AUDIO_OUTPUT/extracted.wav", true)
	if err != nil {
		t.Fatal(err)
	}

	wantPath := filepath.Join(outputRoot, "extracted.wav")
	if destination.path != wantPath || destination.temporary {
		t.Fatalf("resolveAudioDestination(explicit, true) = %+v, want path %q owned by caller", destination, wantPath)
	}

	if _, err := os.Stat(wantPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("explicit destination was created during resolution: %v", err)
	}
}

// TestResolveAudioDestinationCreatesPlaybackTemporaryFile verifies that playback reserves a unique WAV path and marks
// it for command-owned cleanup.
func TestResolveAudioDestinationCreatesPlaybackTemporaryFile(t *testing.T) {
	destination, err := resolveAudioDestination("", true)
	if err != nil {
		t.Fatal(err)
	}
	defer removeTemporaryAudio(destination.path)

	if !destination.temporary {
		t.Fatalf("resolveAudioDestination(\"\", true) = %+v, want temporary ownership", destination)
	}

	if filepath.Ext(destination.path) != ".wav" {
		t.Fatalf("temporary destination %q does not retain the WAV suffix", destination.path)
	}

	if _, err := os.Stat(destination.path); err != nil {
		t.Fatalf("temporary destination was not reserved: %v", err)
	}
}

// TestWriteOrDescribeAudioWithoutDestination verifies the exact inspection message and absence of an output write.
func TestWriteOrDescribeAudioWithoutDestination(t *testing.T) {
	var output bytes.Buffer

	audioData := []byte("wave")

	if err := writeOrDescribeAudio(&output, "sound.wav", audioData, audioDestination{}); err != nil {
		t.Fatal(err)
	}

	if want := "sound.wav: 4 bytes\n"; output.String() != want {
		t.Fatalf("writeOrDescribeAudio output = %q, want %q", output.String(), want)
	}
}

// TestWriteOrDescribeAudioWithDestination verifies that the report follows a successful write and retains its exact
// byte count and host path.
func TestWriteOrDescribeAudioWithDestination(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "extracted.wav")
	destination := audioDestination{path: outputPath}
	audioData := []byte("wave-data")

	var output bytes.Buffer

	if err := writeOrDescribeAudio(&output, "ignored.wav", audioData, destination); err != nil {
		t.Fatal(err)
	}

	written, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(written, audioData) {
		t.Fatalf("written audio = %q, want %q", written, audioData)
	}

	wantOutput := fmt.Sprintf("wrote %d bytes to %s\n", len(audioData), outputPath)
	if output.String() != wantOutput {
		t.Fatalf("writeOrDescribeAudio output = %q, want %q", output.String(), wantOutput)
	}
}

// TestPlayAudioFileRequiresMplayer verifies that dependency discovery fails before command construction and preserves
// the actionable compatibility prefix.
func TestPlayAudioFileRequiresMplayer(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := playAudioFile("unused.wav")
	if !errors.Is(err, exec.ErrNotFound) {
		t.Fatalf("playAudioFile error = %v, want exec.ErrNotFound", err)
	}

	if wantPrefix := "mplayer is required for -play: "; !strings.HasPrefix(err.Error(), wantPrefix) {
		t.Fatalf("playAudioFile error = %q, want prefix %q", err, wantPrefix)
	}
}
