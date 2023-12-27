package cofLoader

import (
	"fmt"
	"io"

	"github.com/gravestench/cof"
)

func (s *Service) Load(filepath string) (*cof.COF, error) {
	stream, err := s.mpq.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data", "error", err)
	}

	cofInstance, err := cof.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("parsing cof", "error", err)
	}

	s.logger.Info("loaded COF", "path", filepath)

	return cofInstance, nil
}

func (s *Service) LoadAnimationData(filepath string) (*cof.AnimationData, error) {
	stream, err := s.mpq.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data", "error", err)
	}

	animData, err := cof.Load(data)
	if err != nil {
		return nil, fmt.Errorf("parsing anim data", "error", err)
	}

	s.logger.Info("loaded Animation Data", "path", filepath)

	return animData, nil
}
