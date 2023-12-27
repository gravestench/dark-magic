package dc6Loader

import (
	"fmt"
	"io"
	"time"

	dc6 "github.com/gravestench/dc6/pkg"
)

func (s *Service) Load(filepath string) (*dc6.DC6, error) {
	for !s.mpq.RequiredArchivesLoaded() {
		time.Sleep(time.Millisecond * 5)
	}

	if s.cache != nil {
		cachedData, isCached := s.cache.Retrieve(filepath)
		if isCached {
			return cachedData.(*dc6.DC6), nil
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

	dc6Image, err := dc6.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("parsing dc6 %q: %v", filepath, err)
	}

	if s.cache != nil {
		if err = s.cache.Insert(filepath, dc6Image, len(data)); err != nil {
			s.logger.Error("caching file", "error", err)
		}
	}

	s.logger.Info("loaded DC6 file", "path", filepath)

	return dc6Image, nil
}
