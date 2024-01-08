package fontTableLoader

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/cache"
	"github.com/gravestench/dark-magic/pkg/services/assetLoader"
	"github.com/gravestench/dark-magic/pkg/services/common"
)

type Service struct {
	common.Service
	*Config
	assets assetLoader.Dependency
	cache  *cache.Cache
}

func (s *Service) Init(mesh servicemesh.Mesh) {

}

func (s *Service) Name() string {
	return "Font Table Loader"
}

func (s *Service) Ready() bool {
	if s.Config == nil {
		return false
	}

	if s.assets == nil {
		return false
	}

	return true
}
