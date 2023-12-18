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

	s.logger.Info("loading", "path", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data", "error", err)
	}

	font, err := font_table.Load(data)
	if err != nil {
		return nil, fmt.Errorf("parsing dt1", "error", err)
	}

	if s.cache != nil {
		if err = s.cache.Insert(filepath, font, len(data)); err != nil {
			s.logger.Error("caching file", "error", err)
		}
	}

	return font, nil
}
