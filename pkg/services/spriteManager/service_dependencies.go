package spriteManager

import (
	"time"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/dc6Loader"
	"github.com/gravestench/dark-magic/pkg/services/dccLoader"
	"github.com/gravestench/dark-magic/pkg/services/ds1Loader"
	"github.com/gravestench/dark-magic/pkg/services/dt1Loader"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/pl2Loader"
)

// the following methods are invoked by the servicemesh
// automatically in an endless loop. As soon as the
// dependencies are resolved, the Init method is called.

func (s *Service) DependenciesResolved() bool {

	if s.dc6 == nil {
		return false
	}

	if s.dcc == nil {
		return false
	}

	if s.ds1 == nil {
		return false
	}

	if s.dt1 == nil {
		return false
	}

	if s.pl2 == nil {
		return false
	}

	if s.mpq == nil {
		return false
	}

	const numDiablo2Archives = 11

	if len(s.mpq.Archives()) != numDiablo2Archives {
		time.Sleep(time.Second)
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(services []servicemesh.Service) {
	for _, service := range services {
		switch candidate := service.(type) {

		case dc6Loader.Dependency:
			s.dc6 = candidate
		case dccLoader.Dependency:
			s.dcc = candidate
		case ds1Loader.Dependency:
			s.ds1 = candidate
		case dt1Loader.Dependency:
			s.dt1 = candidate
		case pl2Loader.Dependency:
			s.pl2 = candidate
		case mpqLoader.Dependency:
			s.mpq = candidate
		}
	}
}
