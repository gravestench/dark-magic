package fontTableLoader

import (
	"fmt"
	"io"

	"github.com/gravestench/font_table"
)

func (s *Service) Load(filepath string) (*font_table.Font, error) {
	if s.cache != nil {
		cachedData, isCached := s.cache.Retrieve(filepath)
		if isCached {
			return cachedData.(*font_table.Font), nil
		}
	}

	s.logger.Info().Msgf("loading %v", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data: %v", err)
	}

	font, err := font_table.Load(data)
	if err != nil {
		return nil, fmt.Errorf("parsing dt1: %v", err)
	}

	if s.cache != nil {
		if err = s.cache.Insert(filepath, font, len(data)); err != nil {
			s.logger.Error().Msgf("caching file %q: %v", err)
		}
	}

	return font, nil
}
