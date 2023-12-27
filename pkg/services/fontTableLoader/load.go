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

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data %q: %v", filepath, err)
	}

	font, err := font_table.Load(data)
	if err != nil {
		return nil, fmt.Errorf("parsing font table %q: %v", filepath, err)
	}

	if s.cache != nil {
		if err = s.cache.Insert(filepath, font, len(data)); err != nil {
			s.logger.Error("caching file", "error", err)
		}
	}

	s.logger.Info("loaded font table", "path", filepath)

	return font, nil
}
