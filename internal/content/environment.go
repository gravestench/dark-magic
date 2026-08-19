package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// standardMPQNames is ordered from newest game data to oldest so patched tables shadow their original definitions.
var standardMPQNames = []string{
	"patch_d2.mpq", "d2exp.mpq", "d2data.mpq", "d2char.mpq",
	"d2music.mpq", "d2sfx.mpq", "d2speech.mpq", "d2video.mpq",
	"d2xmusic.mpq", "d2xtalk.mpq", "d2xvideo.mpq",
}

// FromEnvironment constructs the distribution-supplied mod layers followed by configured external game-data layers.
// Product policy always supplies built-in d2legacy; this generic VFS deliberately does not inject it.
func FromEnvironment(modLayers ...Layer) (*FS, error) {
	layers := make([]Layer, 0, 16)
	layers = append(layers, modLayers...)

	configured := os.Getenv("MPQ_DIRECTORY")
	if configured == "" {
		return New(layers...)
	}

	for index, entry := range strings.Split(configured, ",") {
		var err error

		layers, err = appendMPQDirectoryLayers(layers, index, entry)
		if err != nil {
			return nil, err
		}
	}

	return New(layers...)
}

// appendMPQDirectoryLayers validates one configured root and appends its loose files before its ordered archives.
// Keeping the root together makes the cross-root and within-root priority rules visible at the call site.
func appendMPQDirectoryLayers(layers []Layer, index int, entry string) ([]Layer, error) {
	directory := strings.TrimSpace(entry)
	if directory == "" {
		return nil, fmt.Errorf("content: MPQ_DIRECTORY entry %d is empty", index+1)
	}

	expanded, err := darkpaths.ExpandHost(directory)
	if err != nil {
		return nil, fmt.Errorf("content: expand MPQ directory %q: %w", directory, err)
	}

	directory = expanded

	info, err := os.Stat(directory)
	if err != nil {
		return nil, fmt.Errorf("content: inspect MPQ directory %q: %w", directory, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("content: MPQ path %q is not a directory", directory)
	}

	prefix := fmt.Sprintf("mpq-%d", index)
	layers = append(layers, Layer{Name: prefix + "-directory", FS: Directory(directory)})

	for _, name := range standardMPQNames {
		fileName := filepath.Join(directory, name)
		if _, err := os.Stat(fileName); err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, fmt.Errorf("content: inspect archive %q: %w", fileName, err)
		}

		archive, err := MPQ(fileName)
		if err != nil {
			return nil, err
		}

		layers = append(layers, Layer{Name: prefix + "-" + name, FS: archive})
	}

	return layers, nil
}
