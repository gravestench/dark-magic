package ui

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/common"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

var _ servicemesh.Service = &Service{}

type Service struct {
	common.Service
	renderer raylibRenderer.Dependency
}

func (s *Service) Init(mesh servicemesh.Mesh) {

}

func (s *Service) Name() string {
	return "UI"
}

func (s *Service) Ready() bool {
	if !s.DependenciesResolved() {
		return false
	}

	return true
}
