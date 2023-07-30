package wav_loader

import (
	"io"

	wav "github.com/gravestench/wav/pkg"
)

func (s *Service) Load(filepath string) ([]byte, error) {
	s.logger.Info().Msgf("loading %v", filepath)
	
	stream, err := s.mpq.Load(filepath)
	if err != nil {
		s.logger.Fatal().Msgf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Fatal().Msgf("reading data: %v", err)
	}

	audioData, err := wav.WavDecompress(data, 2)
	if err != nil {
		s.logger.Fatal().Msgf("parsing dt1: %v", err)
	}

	return audioData, nil
}
