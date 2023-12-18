package tblLoader

import (
	"fmt"
	"io"

	tbl "github.com/gravestench/tbl_text"
)

func (s *Service) Load(filepath string) (tbl.TextTable, error) {
	if s.cache != nil {
		cachedData, isCached := s.cache.Retrieve(filepath)
		if isCached {
			return *cachedData.(*tbl.TextTable), nil
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

	table, err := tbl.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("parsing dt1", "error", err)
	}

	if s.cache != nil {
		if err = s.cache.Insert(filepath, table, len(data)); err != nil {
			s.logger.Error("caching file", "error", err)
		}
	}

	return table, nil
}
