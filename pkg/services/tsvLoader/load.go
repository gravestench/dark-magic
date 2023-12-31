package tsvLoader

import (
	"fmt"
	"io"
	"strings"

	"github.com/gravestench/tsv"
)

func (s *Service) Load(filepath string) ([]byte, error) {
	cacheKey := fmt.Sprintf("tsv: %s %s", filepath)

	if s.cache != nil {
		cacheData, isCached := s.cache.Retrieve(cacheKey)
		if isCached {
			return cacheData.([]byte), nil
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

	if s.cache != nil {
		if err = s.cache.Insert(cacheKey, data, len(data)); err != nil {
			s.logger.Warn("caching data", "error", err)
		}
	}

	s.logger.Info("loaded TSV", "path", filepath)

	return data, nil
}

func (s *Service) Unmarshal(filepath string, destination any) error {
	data, err := s.Load(filepath)
	if err != nil {
		return err
	}

	if err = tsv.Unmarshal([]byte(strings.ReplaceAll(string(data), "\"", "")), destination); err != nil {
		return fmt.Errorf("parsing data: %v", err)
	}

	return nil
}
