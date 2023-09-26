package assetLoader

import (
	"time"

	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/cofLoader"
	"github.com/gravestench/dark-magic/pkg/services/dc6Loader"
	"github.com/gravestench/dark-magic/pkg/services/dccLoader"
	"github.com/gravestench/dark-magic/pkg/services/ds1Loader"
	"github.com/gravestench/dark-magic/pkg/services/dt1Loader"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/pl2Loader"
	"github.com/gravestench/dark-magic/pkg/services/tblLoader"
	"github.com/gravestench/dark-magic/pkg/services/tsvLoader"
	"github.com/gravestench/dark-magic/pkg/services/wavLoader"
)

func (s *Service) DependenciesResolved() bool {
	for _, dependency := range []any{
		s.mpq,
		s.dc6,
		s.dcc,
		s.ds1,
		s.dt1,
		s.pl2,
		s.tbl,
		s.tsv,
		s.wav,
		s.cof,
	} {
		if dependency == nil {
			return false
		}
	}

	const numDiablo2Archives = 11

	if len(s.mpq.Archives()) != numDiablo2Archives {
		time.Sleep(time.Second)
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(rt runtime.R) {
	for _, service := range rt.Services() {
		switch candidate := service.(type) {
		case mpqLoader.Dependency:
			s.mpq = candidate
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
		case tblLoader.Dependency:
			s.tbl = candidate
		case tsvLoader.Dependency:
			s.tsv = candidate
		case wavLoader.Dependency:
			s.wav = candidate
		case cofLoader.Dependency:
			s.cof = candidate
		}
	}
}
