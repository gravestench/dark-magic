package main

import (
	"os"

	"github.com/gravestench/dark-magic/internal/presentation/scene"
)

// saveScene truncates and encodes the configured save file while returning the same create or encode failure to
// callers.
func saveScene(state *scene.State, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer closeSceneFile(file)

	return state.Save(file)
}

// loadScene decodes one save file and leaves validation to scene.Load so persisted-format errors remain unchanged.
func loadScene(path string) (*scene.State, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer closeSceneFile(file)

	return scene.Load(file)
}

// closeSceneFile preserves the demo's historical choice to report the primary codec result rather than deferred close
// errors.
func closeSceneFile(file *os.File) {
	_ = file.Close()
}
