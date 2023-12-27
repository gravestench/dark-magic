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

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data %q: %v", filepath, err)
	}

	paletteTransform, err := pl2.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("parsing pl2 %q: %v", filepath, err)
	}

	if s.cache != nil {
		if err = s.cache.Insert(filepath, paletteTransform, len(data)); err != nil {
			s.logger.Error("caching file", "error", err)
		}
	}

	s.logger.Info("loaded PL2 file", "path", filepath)

	return paletteTransform, nil
}
