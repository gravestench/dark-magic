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
		err = fmt.Errorf("loading file %q: %v", filepath, err)
		s.logger.Debug().Msg(err.Error())
		return nil, err
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		err = fmt.Errorf("reading data: %v", err)
		s.logger.Debug().Msg(err.Error())
		return nil, err
	}

	audioData, err := wav.WavDecompress(data, 1)
	if err != nil {
		err = fmt.Errorf("parsing wav: %v", err)
		s.logger.Debug().Msg(err.Error())
		return nil, err
	}

	if s.cache != nil {
		if err = s.cache.Insert(filepath, audioData, len(audioData)); err != nil {
			s.logger.Error().Msgf("caching file %q: %v", err)
		}
	}

	return audioData, nil
}
