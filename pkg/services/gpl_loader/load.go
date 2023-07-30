package gpl_loader

import (
	"bytes"
	"io"

	"github.com/gravestench/gpl"
)

func (s *Service) Load(filepath string) (*gpl.GPL, error) {
	s.logger.Info().Msgf("loading %v", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		s.logger.Fatal().Msgf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Fatal().Msgf("reading data: %v", err)
	}

	palette, err := gpl.Decode(bytes.NewReader(data))
	if err != nil {
		s.logger.Fatal().Msgf("parsing dt1: %v", err)
	}

	return &palette, nil
}
