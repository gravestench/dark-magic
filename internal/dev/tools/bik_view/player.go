package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/gravestench/dark-magic/internal/assets/decode"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// validateAndDescribeBIK verifies the complete payload before exposing its expanded path to an external player.
// Reporting only decoded metadata prevents malformed input from reaching native playback code.
func validateAndDescribeBIK(fileName string, output io.Writer) (string, error) {
	expanded, err := darkpaths.ExpandHost(fileName)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(expanded)
	if err != nil {
		return "", fmt.Errorf("reading BIK: %w", err)
	}

	metadata, err := assetdecode.BIK(data)
	if err != nil {
		return "", err
	}

	// Metadata is informational, so preserve the command's historical best-effort stdout write behavior.
	_, _ = fmt.Fprintf(
		output,
		"%s: %dx%d, %d frames, %.3fs, %d audio track(s)\n",
		metadata.Version,
		metadata.Width,
		metadata.Height,
		metadata.Frames,
		float64(metadata.DurationMillis)/1000,
		len(metadata.AudioTracks),
	)

	return expanded, nil
}

// playBIK resolves the requested executable before attaching it to the process's standard streams.
// Passing one argument slice preserves spaces in paths and avoids shell interpretation of asset names.
func playBIK(playerName, fileName string) error {
	player, err := exec.LookPath(playerName)
	if err != nil {
		return fmt.Errorf("locating %q: %w (install FFmpeg or pass -player)", playerName, err)
	}

	command := exec.Command(player, playerArguments(fileName)...)
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr

	if err := command.Run(); err != nil {
		return fmt.Errorf("running player: %w", err)
	}

	return nil
}

// playerArguments constructs the fixed ffplay-compatible invocation without involving a command shell.
// Keeping the file name as one final element ensures whitespace and option-like characters remain data.
func playerArguments(fileName string) []string {
	return []string{"-autoexit", "-window_title", "Dark Magic BIK Viewer", "-loglevel", "warning", fileName}
}
