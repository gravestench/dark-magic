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

const pcmBlockFrames = 1024

// AttachAudio installs one mixer whose commands and native streams are advanced on the Raylib owner thread.
func (s *Service) AttachAudio(mixer *audio.Mixer) error {
	if mixer == nil {
		return fmt.Errorf("renderer: nil audio mixer")
	}

	s.audioMu.Lock()
	defer s.audioMu.Unlock()

	if s.audioBackend != nil {
		return fmt.Errorf("renderer: audio mixer is already attached")
	}

	backend := newRaylibAudioBackend(mixer)
	s.audioBackend = backend

	// Frame subscription keeps mixer time, command application, and stream updates in their established order.
	s.SubscribeFrame(func() {
		s.runAudioFrame(backend)
	})

	return nil
}

// newRaylibAudioBackend creates all ownership maps together so Apply and Close never need nil-map recovery paths.
func newRaylibAudioBackend(mixer *audio.Mixer) *raylibAudioBackend {
	return &raylibAudioBackend{
		mixer:  mixer,
		sounds: make(map[audio.SoundID]rl.Sound),
		loops:  make(map[audio.SoundID]bool),
		music:  make(map[audio.SoundID]musicPlayback),
		pcm:    make(map[audio.SoundID]*pcmPlayback),
	}
}

// runAudioFrame advances logical time before draining commands, then services native streams created by that drain.
func (s *Service) runAudioFrame(backend *raylibAudioBackend) {
	backend.mixer.Advance(time.Duration(float64(time.Second) * float64(rl.GetFrameTime())))

	if err := backend.mixer.Drain(backend); err != nil && s.logger != nil {
		s.logger.Error("draining audio commands", "error", err)
	}

	backend.Update()
}

// raylibAudioBackend owns every native sound, music stream, and PCM stream created for one attached mixer.
type raylibAudioBackend struct {
	mu     sync.Mutex
	mixer  *audio.Mixer
	sounds map[audio.SoundID]rl.Sound
	loops  map[audio.SoundID]bool
	music  map[audio.SoundID]musicPlayback
	pcm    map[audio.SoundID]*pcmPlayback
}

// pcmPlayback buffers interleaved samples into Raylib-sized blocks and tracks frames acknowledged by the device.
type pcmPlayback struct {
	stream          rl.AudioStream
	channels        int
	pending         []int16
	queued          [][]int16
	started         bool
	stopping        bool
	submittedFrames int
}

// musicPlayback owns the staged encoded source consumed by Raylib's streaming decoder.
// Disk staging avoids retaining multi-megabyte MPQ payloads in the Go heap for the entire frontend session.
type musicPlayback struct {
	music rl.Music
	path  string
}

// Apply serializes command handling with stream updates.
// Holding one lock prevents native handles from being stopped while another frame uses them.
func (b *raylibAudioBackend) Apply(command audio.Command) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch command.Kind {
	case "pcm-open":
		return b.openPCM(command)
	case "pcm-write":
		return b.writePCM(command)
	case "play":
		return b.play(command)
	case "pan":
		return b.setPan(command)
	case "volume":
		return b.setVolume(command)
	case "stop":
		return b.stop(command)
	default:
		return fmt.Errorf("unknown audio command %q", command.Kind)
	}
}

// openPCM temporarily overrides Raylib's global stream buffer size before creating the requested interleaved stream.
func (b *raylibAudioBackend) openPCM(command audio.Command) error {
	if _, exists := b.pcm[command.ID]; exists {
		return fmt.Errorf("PCM stream %v already exists", command.ID)
	}

	rl.SetAudioStreamBufferSizeDefault(pcmBlockFrames)

	stream := rl.LoadAudioStream(uint32(command.Rate), 16, uint32(command.Channels))

	rl.SetAudioStreamBufferSizeDefault(0)

	rl.SetAudioStreamVolume(stream, command.Volume)
	b.pcm[command.ID] = &pcmPlayback{stream: stream, channels: command.Channels}

	return nil
}

// writePCM decodes little-endian samples and queues only complete device blocks, retaining any partial tail.
func (b *raylibAudioBackend) writePCM(command audio.Command) error {
	playback, exists := b.pcm[command.ID]
	if !exists {
		return fmt.Errorf("PCM stream %v does not exist", command.ID)
	}

	playback.pending = append(playback.pending, decodePCM16(command.Data)...)
	playback.queueCompleteBlocks()

	return nil
}

// decodePCM16 converts mixer bytes without changing the established behavior for an odd trailing byte.
func decodePCM16(data []byte) []int16 {
	samples := make([]int16, len(data)/2)
	for index := range samples {
		samples[index] = int16(binary.LittleEndian.Uint16(data[index*2:]))
	}

	return samples
}

// queueCompleteBlocks copies full blocks out of the growable pending buffer so queued slices remain stable.
func (playback *pcmPlayback) queueCompleteBlocks() {
	blockSamples := pcmBlockFrames * playback.channels
	for len(playback.pending) >= blockSamples {
		block := append([]int16(nil), playback.pending[:blockSamples]...)
		playback.queued = append(playback.queued, block)
		playback.pending = playback.pending[blockSamples:]
	}
}

// play rejects duplicate identities before selecting encoded streaming or decoded sound ownership.
func (b *raylibAudioBackend) play(command audio.Command) error {
	if _, exists := b.sounds[command.ID]; exists {
		return fmt.Errorf("sound %v already exists", command.ID)
	}

	if _, exists := b.music[command.ID]; exists {
		return fmt.Errorf("sound %v already exists", command.ID)
	}

	if command.Stream {
		return b.playMusic(command)
	}

	b.playSound(command)

	return nil
}

// playMusic stages encoded bytes before loading the native decoder and removes the file immediately on load failure.
func (b *raylibAudioBackend) playMusic(command audio.Command) error {
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

// playSound unloads the intermediate wave only after Raylib has copied it into a native sound buffer.
func (b *raylibAudioBackend) playSound(command audio.Command) {
	wave := rl.LoadWaveFromMemory(command.Format, command.Data, int32(len(command.Data)))
	sound := rl.LoadSoundFromWave(wave)
	rl.UnloadWave(wave)

	rl.SetSoundVolume(sound, command.Volume)
	rl.SetSoundPan(sound, (command.Pan+1)/2)
	rl.PlaySound(sound)
	b.sounds[command.ID] = sound
	b.loops[command.ID] = command.Loop
}

// setPan prefers music when an identity is present in that ownership map, matching the existing lookup precedence.
func (b *raylibAudioBackend) setPan(command audio.Command) error {
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
}

// setVolume prefers music when an identity is present in that ownership map, matching the existing lookup precedence.
func (b *raylibAudioBackend) setVolume(command audio.Command) error {
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
}

// stop preserves PCM, music, then sound lookup precedence because one command identity may exist across backend kinds.
func (b *raylibAudioBackend) stop(command audio.Command) error {
	if playback, exists := b.pcm[command.ID]; exists {
		playback.queueFinalBlock()
		playback.stopping = true

		return nil
	}

	if playback, exists := b.music[command.ID]; exists {
		b.stopMusic(command.ID, playback)
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
}

// queueFinalBlock zero-pads a partial PCM tail so a stop command drains every supplied sample before unloading.
func (playback *pcmPlayback) queueFinalBlock() {
	if len(playback.pending) == 0 {
		return
	}

	block := make([]int16, pcmBlockFrames*playback.channels)
	copy(block, playback.pending)
	playback.queued = append(playback.queued, block)
	playback.pending = nil
}

// stopMusic releases the native decoder before deleting its staged source file.
func (b *raylibAudioBackend) stopMusic(id audio.SoundID, playback musicPlayback) {
	rl.StopMusicStream(playback.music)
	rl.UnloadMusicStream(playback.music)
	delete(b.music, id)

	_ = os.Remove(playback.path)
}

// Update restarts looping sounds, advances music decoders, and submits PCM in the established backend order.
func (b *raylibAudioBackend) Update() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.updateLoopingSounds()
	b.updateMusic()
	b.updatePCM()
}

// updateLoopingSounds restarts only completed sounds whose command requested manual looping.
func (b *raylibAudioBackend) updateLoopingSounds() {
	for id, loop := range b.loops {
		if loop && !rl.IsSoundPlaying(b.sounds[id]) {
			rl.PlaySound(b.sounds[id])
		}
	}
}

// updateMusic services each streaming decoder once per owner-thread frame.
func (b *raylibAudioBackend) updateMusic() {
	for _, playback := range b.music {
		rl.UpdateMusicStream(playback.music)
	}
}

// updatePCM reports processed frames before submitting the next block, preserving mixer backpressure accounting.
func (b *raylibAudioBackend) updatePCM() {
	for id, playback := range b.pcm {
		processed := rl.IsAudioStreamProcessed(playback.stream)
		if processed && playback.started && playback.submittedFrames > 0 {
			_ = b.mixer.ReportPCMFrames(id, playback.submittedFrames)
			playback.submittedFrames = 0
		}

		if len(playback.queued) > 0 && processed {
			playback.submitNextBlock()
		}

		if playback.finishedStopping() {
			rl.StopAudioStream(playback.stream)
			rl.UnloadAudioStream(playback.stream)
			delete(b.pcm, id)
		}
	}
}

// submitNextBlock retains Raylib-Go's interleaved-slice convention and starts playback only after the first submission.
func (playback *pcmPlayback) submitNextBlock() {
	samples := playback.queued[0]
	// Raylib's C API expects a frame count, while raylib-go derives that count from the slice length. Expose one slice
	// element per frame while retaining the full interleaved backing buffer, matching the existing binding workaround.
	frames := len(samples) / playback.channels
	rl.UpdateAudioStream(playback.stream, samples[:frames])
	playback.queued = playback.queued[1:]

	playback.submittedFrames = frames
	if !playback.started {
		rl.PlayAudioStream(playback.stream)
		playback.started = true
	}
}

// finishedStopping waits for all queued PCM and the final submitted block before native teardown.
func (playback *pcmPlayback) finishedStopping() bool {
	return playback.stopping &&
		len(playback.queued) == 0 &&
		(!playback.started || rl.IsAudioStreamProcessed(playback.stream))
}

// Close tears down sounds, PCM, then music while the native audio device is still available.
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

// stageMusicData preserves a recognized source extension because Raylib selects its decoder from the temporary path.
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
