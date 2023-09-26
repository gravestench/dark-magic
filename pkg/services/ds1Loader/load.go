package ds1Loader

import (
	"io"

	"github.com/gravestench/ds1"
)

func (s *Service) Load(filepath string) (*ds1.DS1, error) {
	if s.cache != nil {
		cachedData, isCached := s.cache.Retrieve(filepath)
		if isCached {
			return cachedData.(*ds1.DS1), nil
		}
	}

	s.logger.Info().Msgf("loading %v", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		s.logger.Fatal().Msgf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Fatal().Msgf("reading data: %v", err)
	}

	ds1Object, err := ds1.FromBytes(data)
	if err != nil {
		s.logger.Fatal().Msgf("parsing ds1: %v", err)
	}

	if s.cache != nil {
		if err = s.cache.Insert(filepath, ds1Object, len(data)); err != nil {
			s.logger.Error().Msgf("caching file '%s': %v", err)
		}
	}

	return ds1Object, nil
}
