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

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data %q: %v", filepath, err)
	}

	dt1Object, err := dt1.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("parsing dt1 %q: %v", filepath, err)
	}

	if s.cache != nil {
		if err = s.cache.Insert(filepath, dt1Object, len(data)); err != nil {
			s.logger.Error("caching file", "error", err)
		}
	}

	s.logger.Info("loaded DT1 file", "path", filepath)

	return dt1Object, nil
}
