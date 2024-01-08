package assetLoader

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/fileLoader"
)

func (s *Service) DependenciesResolved() bool {
	if s.file == nil {
		return false
	}

	//if s.cache.dc6 == nil {
	//	return false
	//}
	//
	//if s.cache.dcc == nil {
	//	return false
	//}
	//
	//if s.cache.ds1 == nil {
	//	return false
	//}
	//
	//if s.cache.dt1 == nil {
	//	return false
	//}
	//
	//if s.cache.cof == nil {
	//	return false
	//}
	//
	//if s.cache.font == nil {
	//	return false
	//}
	//
	//if s.cache.pl2 == nil {
	//	return false
	//}
	//
	//if s.cache.tbl == nil {
	//	return false
	//}
	//
	//if s.cache.tsv == nil {
	//	return false
	//}
	//
	//if s.cache.wav == nil {
	//	return false
	//}

	return true
}

func (s *Service) ResolveDependencies(services []servicemesh.Service) {
	for _, service := range services {
		switch candidate := service.(type) {
		case fileLoader.Dependency:
			s.file = candidate
		}
	}
}
