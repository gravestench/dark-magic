package tsv_loader

import (
	"io"
	"strings"

	"github.com/gravestench/tsv"
)

func (s *Service) Load(filepath string, destination any) error {
	s.logger.Info().Msgf("loading %v", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		s.logger.Fatal().Msgf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Fatal().Msgf("reading data: %v", err)
	}

	if err = tsv.Unmarshal([]byte(strings.ReplaceAll(string(data), "\"", "")), destination); err != nil {
		s.logger.Fatal().Msgf("parsing data: %v", err)
	}

	return nil
}
