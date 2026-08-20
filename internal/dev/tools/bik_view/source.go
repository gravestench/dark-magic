package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gravestench/dark-magic/internal/content"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// standardMPQPriority lists archives from newest overrides to oldest base data.
// The layered filesystem resolves the first matching asset, so changing this order changes the selected movie.
var standardMPQPriority = []string{
	"patch_d2.mpq", "d2exp.mpq", "d2data.mpq", "d2char.mpq",
	"d2music.mpq", "d2sfx.mpq", "d2speech.mpq", "d2video.mpq",
	"d2xmusic.mpq", "d2xtalk.mpq", "d2xvideo.mpq",
}

// resolvePlayableBIK returns a directly playable path and a cleanup function valid after every successful resolution.
// Direct files receive a no-op cleanup so main can retain one ownership path for both input modes.
func resolvePlayableBIK(options commandOptions) (string, func(), error) {
	if options.fileName != "" {
		return options.fileName, func() {}, nil
	}

	return extractBIK(options.sourceName, options.assetName)
}

// extractBIK copies one source asset to a temporary .bik file because external players require a host filesystem path.
// Failed writes remove partial files, while successful callers own the returned idempotent removal function.
func extractBIK(sourceName, assetName string) (string, func(), error) {
	source, err := openSource(sourceName)
	if err != nil {
		return "", func() {}, err
	}
	// Read while the one-shot command owns archive sources; only the extracted host file has explicit cleanup.
	data, err := fs.ReadFile(source, assetName)
	if err != nil {
		return "", func() {}, fmt.Errorf("reading %q: %w", assetName, err)
	}

	file, err := os.CreateTemp("", "dark-magic-*.bik")
	if err != nil {
		return "", func() {}, fmt.Errorf("creating temporary BIK: %w", err)
	}

	fileName := file.Name()
	cleanup := func() { _ = os.Remove(fileName) }

	if _, err := file.Write(data); err != nil {
		_ = file.Close()

		cleanup()

		return "", func() {}, err
	}

	if err := file.Close(); err != nil {
		cleanup()

		return "", func() {}, err
	}

	return fileName, cleanup, nil
}

// openSource normalizes a source path and overlays recognized Diablo II archives found in a directory.
// Non-directories and directories without recognized archives retain content.OpenSource's existing format handling.
func openSource(sourceName string) (fs.FS, error) {
	expanded, err := darkpaths.ExpandHost(sourceName)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(expanded)
	// Delegate stat failures as well as ordinary files so content.OpenSource remains the authority for user-facing errors.
	if err != nil || !info.IsDir() {
		return content.OpenSource(expanded)
	}

	layers, err := openStandardMPQLayers(expanded)
	if err != nil {
		return nil, err
	}

	if len(layers) == 0 {
		return content.OpenSource(expanded)
	}

	return content.New(layers...)
}

// openStandardMPQLayers opens recognized archives in deterministic override order for layered asset lookup.
// Stat failures are skipped for compatibility, while a present but invalid archive still aborts source construction.
func openStandardMPQLayers(directory string) ([]content.Layer, error) {
	layers := make([]content.Layer, 0, len(standardMPQPriority))
	for _, archiveName := range standardMPQPriority {
		archivePath := filepath.Join(directory, archiveName)
		if _, err := os.Stat(archivePath); err != nil {
			continue
		}

		archive, err := content.MPQ(archivePath)
		if err != nil {
			return nil, err
		}

		layers = append(layers, content.Layer{Name: archiveName, FS: archive})
	}

	return layers, nil
}
