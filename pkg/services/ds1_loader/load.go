package ds1_loader

import (
	"io"

	"github.com/gravestench/ds1"
)

func (s *Service) Load(filepath string) (*ds1.DS1, error) {
	s.logger.Info().Msgf("loading %v", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		s.logger.Fatal().Msgf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Fatal().Msgf("reading data: %v", err)
	}

	ds1Object, err := ds1.FromBytes(data)
	if err != nil {
		s.logger.Fatal().Msgf("parsing ds1: %v", err)
	}

	return ds1Object, nil
}
