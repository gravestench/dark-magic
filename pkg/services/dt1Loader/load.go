package dt1Loader

import (
	"fmt"
	"io"

	"github.com/gravestench/dt1"
)

func (s *Service) Load(filepath string) (*dt1.DT1, error) {
	if s.cache != nil {
		cachedData, isCached := s.cache.Retrieve(filepath)
		if isCached {
			return cachedData.(*dt1.DT1), nil
		}
	}

	s.logger.Info().Msgf("loading %v", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		s.logger.Error().Msgf("loading file %q: %v", filepath, err)
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Error().Msgf("reading data: %v", err)
		return nil, fmt.Errorf("reading data: %v", err)
	}

	dt1Object, err := dt1.FromBytes(data)
	if err != nil {
		s.logger.Error().Msgf("parsing dt1: %v", err)
		return nil, fmt.Errorf("parsing dt1: %v", err)
	}

	if s.cache != nil {
		if err = s.cache.Insert(filepath, dt1Object, len(data)); err != nil {
			s.logger.Error().Msgf("caching file '%s': %v", err)
		}
	}

	return dt1Object, nil
}
