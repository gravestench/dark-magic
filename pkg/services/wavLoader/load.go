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

	s.logger.Info().Msgf("loading %v", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data: %v", err)
	}

	audioData, err := wav.WavDecompress(data, 1)
	if err != nil {
		return nil, fmt.Errorf("parsing wav: %v", err)
	}

	if s.cache != nil {
		if err = s.cache.Insert(filepath, audioData, len(audioData)); err != nil {
			s.logger.Error().Msgf("caching file %q: %v", err)
		}
	}

	return audioData, nil
}
