package goscript

import (
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func (s *Service) runScript(scriptPath string) {
	s.logger.Info("running script", "path", scriptPath)
	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)
	//i.Use(darkapi.Symbols)
	if _, err := i.EvalPath(scriptPath); err != nil {
		s.logger.Warn("executing init script", "path", scriptPath, "error", err)
	}
}
