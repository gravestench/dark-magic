package dt1_loader

import (
	"io"

	"github.com/gravestench/dt1"
)

func (s *Service) Load(filepath string) (*dt1.DT1, error) {
	s.logger.Info().Msgf("loading %v", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		s.logger.Fatal().Msgf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Fatal().Msgf("reading data: %v", err)
	}

	dt1Object, err := dt1.FromBytes(data)
	if err != nil {
		s.logger.Fatal().Msgf("parsing dt1: %v", err)
	}

	return dt1Object, nil
}
