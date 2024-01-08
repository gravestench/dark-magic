package goscript

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/common"
)

type Service struct {
	common.Service
	*Config
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.runScript(s.Config.InitScriptPath)

	// TODO: add file watcher for re-running init script.
}

func (s *Service) Name() string {
	return "Goscript"
}

func (s *Service) Ready() bool {
	if s.Config == nil {
		return false
	}

	return true
}
