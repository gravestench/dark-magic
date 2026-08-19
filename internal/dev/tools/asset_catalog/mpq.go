package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gravestench/dark-magic/internal/content"
)

// openMPQStack opens every supported archive in overlay priority order. Missing optional archives are tolerated, while
// other filesystem or archive errors stop construction before callers receive an incomplete stack.
func openMPQStack(directory string) (*content.FS, error) {
	priority := []string{
		"patch_d2.mpq", "d2exp.mpq", "d2data.mpq", "d2char.mpq",
		"d2music.mpq", "d2sfx.mpq", "d2speech.mpq", "d2video.mpq",
		"d2xmusic.mpq", "d2xtalk.mpq", "d2xvideo.mpq",
	}
	layers := make([]content.Layer, 0, len(priority))

	// Iteration order is the content overlay contract: earlier archives must retain lookup precedence.
	for _, name := range priority {
		path := filepath.Join(directory, name)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, err
		}

		archive, err := content.OpenSource(path)
		if err != nil {
			return nil, err
		}

		layers = append(layers, content.Layer{Name: name, FS: archive})
	}

	if len(layers) == 0 {
		return nil, fmt.Errorf("no supported MPQs found in %q", directory)
	}

	return content.New(layers...)
}
