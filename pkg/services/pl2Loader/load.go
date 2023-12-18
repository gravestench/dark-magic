package pl2Loader

import (
	"fmt"
	"io"

	"github.com/gravestench/pl2"
)

func (s *Service) Load(filepath string) (*pl2.PL2, error) {
	if s.cache != nil {
		cachedData, isCached := s.cache.Retrieve(filepath)
		if isCached {
			return cachedData.(*pl2.PL2), nil
		}
	}

	s.logger.Info("loading", "path", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file", "error", err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data", "error", err)
	}

	paletteTransform, err := pl2.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("parsing dt1", "error", err)
	}

	if s.cache != nil {
		if err = s.cache.Insert(filepath, paletteTransform, len(data)); err != nil {
			s.logger.Error("caching file", "error", err)
		}
	}

	return paletteTransform, nil
}
