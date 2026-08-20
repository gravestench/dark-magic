package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gravestench/dark-magic/internal/content"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

type audioDestination struct {
	path      string
	temporary bool
}

// readLayeredAudioAsset resolves bytes through the same environment-composed content stack used by the application,
// so directory and archive priority remain authoritative.
func readLayeredAudioAsset(assetPath string) ([]byte, error) {
	contentFS, err := content.FromEnvironment()
	if err != nil {
		return nil, err
	}

	return fs.ReadFile(contentFS, assetPath)
}

// resolveAudioDestination expands an explicit host path or creates the playback-only temporary file that reserves a
// collision-free name before data is written.
func resolveAudioDestination(requestedPath string, play bool) (audioDestination, error) {
	if requestedPath != "" {
		expandedPath, err := darkpaths.ExpandHost(requestedPath)
		if err != nil {
			return audioDestination{}, err
		}

		return audioDestination{path: expandedPath}, nil
	}

	if !play {
		return audioDestination{}, nil
	}

	temporaryFile, err := os.CreateTemp("", "dark-magic-*.wav")
	if err != nil {
		return audioDestination{}, err
	}

	// The later write reopens this path, so retaining the descriptor would only extend ownership and leak it into
	// mplayer. Close errors remain non-fatal for compatibility with the original short-lived command.
	temporaryPath := temporaryFile.Name()
	_ = temporaryFile.Close()

	return audioDestination{path: temporaryPath, temporary: true}, nil
}

// writeOrDescribeAudio writes requested output before reporting it, while the no-output path reports only the source
// asset size and leaves the host filesystem untouched.
func writeOrDescribeAudio(output io.Writer, assetPath string, audioData []byte, destination audioDestination) error {
	if destination.path == "" {
		// Status output is best effort so a closed diagnostic stream cannot turn a successful inspection into failure.
		_, _ = fmt.Fprintf(output, "%s: %d bytes\n", assetPath, len(audioData))

		return nil
	}

	if err := os.WriteFile(destination.path, audioData, 0o644); err != nil {
		return err
	}

	// The file write is authoritative; failure to print its confirmation does not invalidate the completed output.
	_, _ = fmt.Fprintf(output, "wrote %d bytes to %s\n", len(audioData), destination.path)

	return nil
}

// removeTemporaryAudio performs best-effort cleanup because the command cannot recover from a removal failure.
func removeTemporaryAudio(path string) {
	_ = os.Remove(path)
}

// playAudioFile validates mplayer before starting it and attaches the child streams directly to the command's streams
// so playback diagnostics and ordering remain unchanged.
func playAudioFile(audioPath string) error {
	if _, err := exec.LookPath("mplayer"); err != nil {
		return fmt.Errorf("mplayer is required for -play: %w", err)
	}

	command := exec.Command("mplayer", filepath.Clean(audioPath))
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	return command.Run()
}
