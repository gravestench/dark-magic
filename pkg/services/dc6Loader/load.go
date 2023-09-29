package dc6Loader

import (
	"io"

	"github.com/gravestench/dc6"
)

func (s *Service) Load(filepath string) (*dc6.DC6, error) {
	if s.cache != nil {
		cachedData, isCached := s.cache.Retrieve(filepath)
		if isCached {
			return cachedData.(*dc6.DC6), nil
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

	dc6Image, err := dc6.FromBytes(data)
	if err != nil {
		s.logger.Fatal().Msgf("parsing dc6: %v", err)
	}

	if s.cache != nil {
		if err = s.cache.Insert(filepath, dc6Image, len(data)); err != nil {
			s.logger.Error().Msgf("caching file %q: %v", err)
		}
	}

	return dc6Image, nil
}
