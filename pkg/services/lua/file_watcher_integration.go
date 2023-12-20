package lua

import (
	"github.com/gravestench/dark-magic/pkg/services/fileWatcher"
)

var _ fileWatcher.NeedsFileWatcher = &Service{}

func (s *Service) FileHandlers() map[string]fileWatcher.FileHandlerFunc {
	handlers := make(map[string]fileWatcher.FileHandlerFunc)

	initScriptPath := s.config.Group(s.Name()).GetString("init script")

	handlers[initScriptPath] = func(path string) error {
		return s.runScript(path)
	}

	return handlers
}
