package backgroundMusic

import "C"
import (
	"log/slog"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/paths"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
	"github.com/gravestench/dark-magic/pkg/services/wavLoader"
)

var _ BackgroundMusicManager = &Service{}

type Service struct {
	logger     *slog.Logger
	mpqWavLoad wavLoader.Dependency
	renderer   raylibRenderer.Dependency
	config     *configFile.Config
}

func (s *Service) DependenciesResolved() bool {
	if s.mpqWavLoad == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(services []servicemesh.Service) {
	for _, service := range services {
		switch candidate := service.(type) {
		case wavLoader.Dependency:
			s.mpqWavLoad = candidate
		case raylibRenderer.Dependency:
			s.renderer = candidate
		}
	}
}

func (s *Service) SetLogger(l *slog.Logger) {
	s.logger = l
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	path := paths.BGMTitle
	s.logger.Info("playing music", "path", path)
	if err := s.playAudio(path); err != nil {
		s.logger.Error("playing music", "path", path, "error", err)
	}
}

func (s *Service) Name() string {
	return "Background Music"
}

func (s *Service) Ready() bool {
	if !s.DependenciesResolved() {
		return false
	}

	if !s.mpqWavLoad.Ready() {
		return false
	}

	if s.renderer == nil {
		return false
	}

	if !s.renderer.IsInit() {
		return false
	}

	return true
}

func (s *Service) SetBackgroundMusic(path string) error {
	return nil
}

func (s *Service) playAudio(path string) error {
	wavBytes, err := s.mpqWavLoad.Load(path)
	if err != nil {
		return err
	}

	const (
		sampleRate  = 44100 / 2
		sampleSize  = 16
		numChannels = 2
	)

	numBytes := len(wavBytes)
	sampleCount := uint32(numBytes / numChannels)
	wavStream := rl.NewWave(sampleCount, sampleRate, sampleSize, numChannels, wavBytes)
	rlSound := rl.LoadSoundFromWave(wavStream)

	volume := s.config.Group(groupBGM).GetFloat(keyBGMVolume)

	rl.SetMasterVolume(0.0)
	rl.PlaySound(rlSound)
	time.Sleep(time.Millisecond * 100)
	rl.SetMasterVolume(float32(volume))

	return nil
}
