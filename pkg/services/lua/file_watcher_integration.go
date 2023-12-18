package lua

import (
	"github.com/gravestench/dark-magic/pkg/services/fileWatcher"
)

var _ fileWatcher.NeedsFileWatcher = &Service{}

func (s *Service) FileHandlers() map[string]fileWatcher.FileHandlerFunc {
	handlers := make(map[string]fileWatcher.FileHandlerFunc)

	cfg, err := s.cfg.GetConfigByFileName(s.ConfigFileName())
	if err != nil {
		s.logger.Error("getting config", "error", err)
		return handlers
	}

	initScriptPath := cfg.Group(s.Name()).GetString("init script")

	handlers[initScriptPath] = func(path string) error {
		return s.runScript(path)
	}

	return handlers
}
