package cofLoader

import (
	"io"

	"github.com/gravestench/cof"
)

func (s *Service) Load(filepath string) (*cof.COF, error) {
	s.logger.Info().Msgf("loading %v", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		s.logger.Fatal().Msgf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Fatal().Msgf("reading data: %v", err)
	}

	cofInstance, err := cof.Unmarshal(data)
	if err != nil {
		s.logger.Fatal().Msgf("parsing cof: %v", err)
	}

	return cofInstance, nil
}

func (s *Service) LoadAnimationData(filepath string) (*cof.AnimationData, error) {
	s.logger.Info().Msgf("loading %v", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		s.logger.Fatal().Msgf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Fatal().Msgf("reading data: %v", err)
	}

	animData, err := cof.Load(data)
	if err != nil {
		s.logger.Fatal().Msgf("parsing cof: %v", err)
	}

	return animData, nil
}
