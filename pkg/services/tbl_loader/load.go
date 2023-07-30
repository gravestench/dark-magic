package tbl_loader

import (
	"io"

	tbl "github.com/gravestench/tbl_text"
)

func (s *Service) Load(filepath string) (*tbl.TextTable, error) {
	s.logger.Info().Msgf("loading %v", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		s.logger.Fatal().Msgf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Fatal().Msgf("reading data: %v", err)
	}

	table, err := tbl.Unmarshal(data)
	if err != nil {
		s.logger.Fatal().Msgf("parsing dt1: %v", err)
	}

	return &table, nil
}
