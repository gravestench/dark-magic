package pl2Loader

import (
	"fmt"
	"io"

	"github.com/gravestench/pl2"
)

func (s *Service) Load(filepath string) (*pl2.PL2, error) {
	s.logger.Info().Msgf("loading %v", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		s.logger.Error().Msgf("loading file %q: %v", filepath, err)
		return nil, fmt.Errorf("loading file: %v", err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Error().Msgf("reading data: %v", err)
		return nil, fmt.Errorf("reading data: %v", err)
	}

	paletteTransform, err := pl2.FromBytes(data)
	if err != nil {
		s.logger.Error().Msgf("parsing dt1: %v", err)
		return nil, fmt.Errorf("parsing dt1: %v", err)
	}

	return paletteTransform, nil
}
