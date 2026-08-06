package raylibRenderer

import (
	"fmt"
	"sync"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/gravestench/dark-magic/internal/audiocore"
)

// AttachAudio drains audio commands on the Raylib owner thread.
func (s *Service) AttachAudio(mixer *audiocore.Mixer) error {
	if mixer == nil {
		return fmt.Errorf("renderer: nil audio mixer")
	}
	s.audioMu.Lock()
	defer s.audioMu.Unlock()
	if s.audioBackend != nil {
		return fmt.Errorf("renderer: audio mixer is already attached")
	}
	backend := &raylibAudioBackend{sounds: make(map[audiocore.SoundID]rl.Sound), loops: make(map[audiocore.SoundID]bool)}
	s.audioBackend = backend
	s.SubscribeFrame(func() {
		mixer.Advance(time.Duration(float64(time.Second) * float64(rl.GetFrameTime())))
		if err := mixer.Drain(backend); err != nil && s.logger != nil {
			s.logger.Error("draining audio commands", "error", err)
		}
		backend.Update()
	})
	return nil
}

type raylibAudioBackend struct {
	mu     sync.Mutex
	sounds map[audiocore.SoundID]rl.Sound
	loops  map[audiocore.SoundID]bool
}

func (b *raylibAudioBackend) Apply(command audiocore.Command) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch command.Kind {
	case "play":
		if _, exists := b.sounds[command.ID]; exists {
			return fmt.Errorf("sound %v already exists", command.ID)
		}
		wave := rl.LoadWaveFromMemory(command.Format, command.Data, int32(len(command.Data)))
		sound := rl.LoadSoundFromWave(wave)
		rl.UnloadWave(wave)
		rl.SetSoundVolume(sound, command.Volume)
		rl.SetSoundPan(sound, (command.Pan+1)/2)
		rl.PlaySound(sound)
		b.sounds[command.ID] = sound
		b.loops[command.ID] = command.Loop
		return nil
	case "pan":
		sound, exists := b.sounds[command.ID]
		if !exists {
			return fmt.Errorf("sound %v does not exist", command.ID)
		}
		rl.SetSoundPan(sound, (command.Pan+1)/2)
		return nil
	case "volume":
		sound, exists := b.sounds[command.ID]
		if !exists {
			return fmt.Errorf("sound %v does not exist", command.ID)
		}
		rl.SetSoundVolume(sound, command.Volume)
		return nil
	case "stop":
		sound, exists := b.sounds[command.ID]
		if !exists {
			return fmt.Errorf("sound %v does not exist", command.ID)
		}
		rl.StopSound(sound)
		rl.UnloadSound(sound)
		delete(b.sounds, command.ID)
		delete(b.loops, command.ID)
		return nil
	default:
		return fmt.Errorf("unknown audio command %q", command.Kind)
	}
}

func (b *raylibAudioBackend) Update() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, loop := range b.loops {
		if loop && !rl.IsSoundPlaying(b.sounds[id]) {
			rl.PlaySound(b.sounds[id])
		}
	}
}

func (b *raylibAudioBackend) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, sound := range b.sounds {
		rl.StopSound(sound)
		rl.UnloadSound(sound)
		delete(b.sounds, id)
	}
}
