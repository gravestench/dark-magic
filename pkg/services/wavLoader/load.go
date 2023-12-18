package wavLoader

import (
	"fmt"
	"io"

	wav "github.com/gravestench/wav/pkg"
)

func (s *Service) Load(filepath string) ([]byte, error) {
	if s.cache != nil {
		cachedData, isCached := s.cache.Retrieve(filepath)
		if isCached {
			return cachedData.([]byte), nil
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

	audioData, err := wav.WavDecompress(data, 1)
	if err != nil {
		return nil, fmt.Errorf("parsing wav", "error", err)
	}

	if s.cache != nil {
		if err = s.cache.Insert(filepath, audioData, len(audioData)); err != nil {
			s.logger.Error("caching file", "error", err)
		}
	}

	return audioData, nil
}
