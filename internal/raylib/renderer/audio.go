package raylibRenderer

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/gravestench/dark-magic/internal/audio"
)

// AttachAudio drains audio commands on the Raylib owner thread.
func (s *Service) AttachAudio(mixer *audio.Mixer) error {
	if mixer == nil {
		return fmt.Errorf("renderer: nil audio mixer")
	}
	s.audioMu.Lock()
	defer s.audioMu.Unlock()
	if s.audioBackend != nil {
		return fmt.Errorf("renderer: audio mixer is already attached")
	}
	backend := &raylibAudioBackend{mixer: mixer, sounds: make(map[audio.SoundID]rl.Sound), loops: make(map[audio.SoundID]bool), music: make(map[audio.SoundID]musicPlayback), pcm: make(map[audio.SoundID]*pcmPlayback)}
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
	mixer  *audio.Mixer
	sounds map[audio.SoundID]rl.Sound
	loops  map[audio.SoundID]bool
	music  map[audio.SoundID]musicPlayback
	pcm    map[audio.SoundID]*pcmPlayback
}

type pcmPlayback struct {
	stream          rl.AudioStream
	channels        int
	pending         []int16
	queued          [][]int16
	started         bool
	stopping        bool
	submittedFrames int
}

// musicPlayback owns the staged encoded source consumed by raylib's streaming
// decoder. Keeping it on disk avoids retaining multi-megabyte MPQ payloads in
// the Go heap for the entire frontend session.
type musicPlayback struct {
	music rl.Music
	path  string
}

const pcmBlockFrames = 1024

func (b *raylibAudioBackend) Apply(command audio.Command) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch command.Kind {
	case "pcm-open":
		if _, exists := b.pcm[command.ID]; exists {
			return fmt.Errorf("PCM stream %v already exists", command.ID)
		}
		rl.SetAudioStreamBufferSizeDefault(pcmBlockFrames)
		stream := rl.LoadAudioStream(uint32(command.Rate), 16, uint32(command.Channels))
		rl.SetAudioStreamBufferSizeDefault(0)
		rl.SetAudioStreamVolume(stream, command.Volume)
		b.pcm[command.ID] = &pcmPlayback{stream: stream, channels: command.Channels}
		return nil
	case "pcm-write":
		playback, exists := b.pcm[command.ID]
		if !exists {
			return fmt.Errorf("PCM stream %v does not exist", command.ID)
		}
		samples := make([]int16, len(command.Data)/2)
		for index := range samples {
			samples[index] = int16(binary.LittleEndian.Uint16(command.Data[index*2:]))
		}
		playback.pending = append(playback.pending, samples...)
		blockSamples := pcmBlockFrames * playback.channels
		for len(playback.pending) >= blockSamples {
			block := append([]int16(nil), playback.pending[:blockSamples]...)
			playback.queued = append(playback.queued, block)
			playback.pending = playback.pending[blockSamples:]
		}
		return nil
	case "play":
		if _, exists := b.sounds[command.ID]; exists {
			return fmt.Errorf("sound %v already exists", command.ID)
		}
		if _, exists := b.music[command.ID]; exists {
			return fmt.Errorf("sound %v already exists", command.ID)
		}
		if command.Stream {
			path, err := stageMusicData(command.Format, command.Data)
			if err != nil {
				return err
			}
			music := rl.LoadMusicStream(path)
			if music.CtxData == nil {
				_ = os.Remove(path)
				return fmt.Errorf("loading staged music %q", command.Format)
			}
			music.Looping = command.Loop
			rl.SetMusicVolume(music, command.Volume)
			rl.SetMusicPan(music, (command.Pan+1)/2)
			rl.PlayMusicStream(music)
			b.music[command.ID] = musicPlayback{music: music, path: path}
			return nil
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
		if playback, exists := b.music[command.ID]; exists {
			rl.SetMusicPan(playback.music, (command.Pan+1)/2)
			return nil
		}
		sound, exists := b.sounds[command.ID]
		if !exists {
			return fmt.Errorf("sound %v does not exist", command.ID)
		}
		rl.SetSoundPan(sound, (command.Pan+1)/2)
		return nil
	case "volume":
		if playback, exists := b.music[command.ID]; exists {
			rl.SetMusicVolume(playback.music, command.Volume)
			return nil
		}
		sound, exists := b.sounds[command.ID]
		if !exists {
			return fmt.Errorf("sound %v does not exist", command.ID)
		}
		rl.SetSoundVolume(sound, command.Volume)
		return nil
	case "stop":
		if playback, exists := b.pcm[command.ID]; exists {
			if len(playback.pending) > 0 {
				blockSamples := pcmBlockFrames * playback.channels
				block := make([]int16, blockSamples)
				copy(block, playback.pending)
				playback.queued = append(playback.queued, block)
				playback.pending = nil
			}
			playback.stopping = true
			return nil
		}
		if playback, exists := b.music[command.ID]; exists {
			rl.StopMusicStream(playback.music)
			rl.UnloadMusicStream(playback.music)
			delete(b.music, command.ID)
			_ = os.Remove(playback.path)
			return nil
		}
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
	for _, playback := range b.music {
		rl.UpdateMusicStream(playback.music)
	}
	for id, playback := range b.pcm {
		processed := rl.IsAudioStreamProcessed(playback.stream)
		if processed && playback.started && playback.submittedFrames > 0 {
			_ = b.mixer.ReportPCMFrames(id, playback.submittedFrames)
			playback.submittedFrames = 0
		}
		if len(playback.queued) > 0 && processed {
			samples := playback.queued[0]
			// raylib's C API expects a frame count, while raylib-go derives that
			// count directly from the slice length. For interleaved audio, expose
			// one element per frame while retaining the complete backing buffer.
			frames := len(samples) / playback.channels
			rl.UpdateAudioStream(playback.stream, samples[:frames])
			playback.queued = playback.queued[1:]
			playback.submittedFrames = frames
			if !playback.started {
				rl.PlayAudioStream(playback.stream)
				playback.started = true
			}
		}
		if playback.stopping && len(playback.queued) == 0 && (!playback.started || rl.IsAudioStreamProcessed(playback.stream)) {
			rl.StopAudioStream(playback.stream)
			rl.UnloadAudioStream(playback.stream)
			delete(b.pcm, id)
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
	for id, playback := range b.pcm {
		rl.StopAudioStream(playback.stream)
		rl.UnloadAudioStream(playback.stream)
		delete(b.pcm, id)
	}
	for id, playback := range b.music {
		rl.StopMusicStream(playback.music)
		rl.UnloadMusicStream(playback.music)
		delete(b.music, id)
		_ = os.Remove(playback.path)
	}
}

func stageMusicData(format string, data []byte) (string, error) {
	extension := strings.ToLower(filepath.Ext(format))
	if extension == "" {
		extension = ".audio"
	}
	file, err := os.CreateTemp("", "darkmagic-music-*"+extension)
	if err != nil {
		return "", fmt.Errorf("staging music: %w", err)
	}
	path := file.Name()
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("staging music payload: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("closing staged music: %w", err)
	}
	return path, nil
}
