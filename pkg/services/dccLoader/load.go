package dccLoader

import (
	"fmt"
	"io"

	"github.com/gravestench/dcc"
)

func (s *Service) Load(filepath string) (*dcc.DCC, error) {
	if s.cache != nil {
		cachedData, isCached := s.cache.Retrieve(filepath)
		if isCached {
			return cachedData.(*dcc.DCC), nil
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

	dccImage, err := dcc.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("parsing dcc %q: %v", filepath, err)
	}

	if s.cache != nil {
		if err = s.cache.Insert(filepath, dccImage, len(data)); err != nil {
			s.logger.Error("caching file", "error", err)
		}
	}

	s.logger.Info("loaded DCC file", "path", filepath)

	return dccImage, nil
}
