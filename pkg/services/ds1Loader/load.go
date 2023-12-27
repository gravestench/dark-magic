package ds1Loader

import (
	"fmt"
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

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data %q: %v", filepath, err)
	}

	ds1Object, err := ds1.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("parsing ds1 %q: %v", filepath, err)
	}

	if s.cache != nil {
		if err = s.cache.Insert(filepath, ds1Object, len(data)); err != nil {
			s.logger.Error("caching file", "error", err)
		}
	}

	s.logger.Info("loaded DS1 file", "path", filepath)

	return ds1Object, nil
}
