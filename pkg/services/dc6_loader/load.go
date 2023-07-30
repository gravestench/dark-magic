package dc6_loader

import (
	"io"

	"github.com/gravestench/dc6"
)

func (s *Service) Load(filepath string) (*dc6.DC6, error) {
	s.logger.Info().Msgf("loading %v", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		s.logger.Fatal().Msgf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Fatal().Msgf("reading data: %v", err)
	}

	dc6Image, err := dc6.FromBytes(data)
	if err != nil {
		s.logger.Fatal().Msgf("parsing dc6: %v", err)
	}

	return dc6Image, nil
}
