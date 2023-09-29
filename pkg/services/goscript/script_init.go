package goscript

import (
	"github.com/gravestench/dark-magic/pkg/darkapi"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func (s *Service) runScript(scriptPath string) {
	s.logger.Info().Msgf("running script %q", scriptPath)
	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)
	i.Use(darkapi.Symbols)
	if _, err := i.EvalPath(scriptPath); err != nil {
		s.logger.Warn().Msgf("executing init script '%s': %v", scriptPath, err)
	}
}
