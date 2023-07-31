package wav_loader

import (
	"fmt"
	"io"

	wav "github.com/gravestench/wav/pkg"
)

func (s *Service) Load(filepath string) ([]byte, error) {
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

	return audioData, nil
}
