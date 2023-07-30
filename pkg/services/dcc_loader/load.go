package dcc_loader

import (
	"io"

	"github.com/gravestench/dcc"
)

func (s *Service) Load(filepath string) (*dcc.DCC, error) {
	s.logger.Info().Msgf("loading %v", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		s.logger.Fatal().Msgf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Fatal().Msgf("reading data: %v", err)
	}

	dccImage, err := dcc.FromBytes(data)
	if err != nil {
		s.logger.Fatal().Msgf("parsing dcc: %v", err)
	}

	return dccImage, nil
}
