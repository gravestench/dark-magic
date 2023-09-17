package fontTableLoader

import (
	"io"

	"github.com/gravestench/font_table"
)

func (s *Service) Load(filepath string) (*font_table.Font, error) {
	s.logger.Info().Msgf("loading %v", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		s.logger.Fatal().Msgf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Fatal().Msgf("reading data: %v", err)
	}

	font, err := font_table.Load(data)
	if err != nil {
		s.logger.Fatal().Msgf("parsing dt1: %v", err)
	}

	return font, nil
}
