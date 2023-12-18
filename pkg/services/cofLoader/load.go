package cofLoader

import (
	"io"

	"github.com/gravestench/cof"
)

func (s *Service) Load(filepath string) (*cof.COF, error) {
	s.logger.Info("loading", "path", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		s.logger.Error("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Error("reading data", "error", err)
	}

	cofInstance, err := cof.Unmarshal(data)
	if err != nil {
		s.logger.Error("parsing cof", "error", err)
	}

	return cofInstance, nil
}

func (s *Service) LoadAnimationData(filepath string) (*cof.AnimationData, error) {
	s.logger.Info("loading", "path", filepath)

	stream, err := s.mpq.Load(filepath)
	if err != nil {
		s.logger.Error("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		s.logger.Error("reading data", "error", err)
	}

	animData, err := cof.Load(data)
	if err != nil {
		s.logger.Error("parsing cof", "error", err)
	}

	return animData, nil
}
