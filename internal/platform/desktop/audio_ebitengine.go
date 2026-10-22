//go:build ebitengine

package desktop

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"sync"

	ebitaudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/cache"
)

const (
	ebitAudioRate           = 48000
	ebitSampleCacheBytes    = 16 * 1024 * 1024
	ebitMaxSampleBytes      = 1024 * 1024
	ebitDecoderReserveBytes = 512 * 1024
)

type ebitAudioPlayer interface {
	Play()
	IsPlaying() bool
	SetVolume(float64)
	Close() error
}

type ebitVoice struct {
	player ebitAudioPlayer
	pan    *stereoPanReader
}

type ebitSample struct {
	mu      sync.Mutex
	decoder io.Reader
	data    []byte
	length  int64
	err     error
}

type ebitSampleReader struct {
	sample   *ebitSample
	position int64
	loop     bool
}

func (s *ebitSample) reader(loop bool) *ebitSampleReader {
	return &ebitSampleReader{sample: s, loop: loop}
}

// Read publishes immutable PCM blocks on the requesting audio reader's thread.
// The sample lock serializes decoding, while each player owns its own cursor.
func (r *ebitSampleReader) Read(output []byte) (int, error) {
	if len(output) == 0 {
		return 0, nil
	}

	s := r.sample
	s.mu.Lock()
	defer s.mu.Unlock()

	limit := int64(cap(s.data))
	if r.loop {
		limit = s.length
	}

	if r.position >= limit {
		return 0, io.EOF
	}

	target := r.position + min(int64(len(output)), limit-r.position)
	for int64(len(s.data)) < target && s.err == nil {
		// Match the streaming pan reader's blocks, including one-shot decoder padding.
		end := min(len(s.data)+4096, cap(s.data))
		n, err := io.ReadFull(s.decoder, s.data[len(s.data):end])

		s.data = s.data[:len(s.data)+n]
		if err == nil && int64(len(s.data)) >= s.length {
			err = io.EOF
		}

		if err != nil {
			if err == io.ErrUnexpectedEOF {
				err = io.EOF
			}

			s.err = err
			s.decoder = nil
		}
	}

	if r.position >= int64(len(s.data)) {
		return 0, s.err
	}

	n := copy(output, s.data[r.position:min(target, int64(len(s.data)))])
	r.position += int64(n)

	return n, nil
}

func (r *ebitSampleReader) Seek(offset int64, whence int) (int64, error) {
	base := int64(0)

	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		base = r.position
	case io.SeekEnd:
		base = r.sample.length
	default:
		return 0, fmt.Errorf("invalid audio seek origin %d", whence)
	}

	position := base + offset
	if position < 0 || position > (r.sample.length/4096+1)*4096 {
		return 0, fmt.Errorf("audio seek out of range")
	}

	r.position = position

	return position, nil
}

type ebitAudioBackend struct {
	mixer     *audio.Mixer
	logger    *slog.Logger
	voices    map[audio.SoundID]ebitVoice
	finished  []audio.SoundID
	samples   *cache.Cache
	newPlayer func(io.Reader) (ebitAudioPlayer, error)
}

func newEbitAudioBackend(mixer *audio.Mixer, logger *slog.Logger) *ebitAudioBackend {
	return &ebitAudioBackend{
		mixer: mixer, logger: logger, voices: make(map[audio.SoundID]ebitVoice),
		samples: cache.New(ebitSampleCacheBytes),
		newPlayer: func(source io.Reader) (ebitAudioPlayer, error) {
			context := ebitaudio.CurrentContext()
			if context == nil {
				context = ebitaudio.NewContext(ebitAudioRate)
			}
			return context.NewPlayer(source)
		},
	}
}

func (b *ebitAudioBackend) Apply(command audio.Command) error {
	voice, exists := b.voices[command.ID]
	switch command.Kind {
	case "play":
		if exists {
			return fmt.Errorf("ebitengine audio: duplicate sound %v", command.ID)
		}
		if err := b.play(command); err != nil {
			// Permanent asset errors must not hold the mixer's retry queue ahead of valid cues.
			b.logger.Error("playing Ebitengine sound", "sound", command.ID, "error", err)
			b.finished = append(b.finished, command.ID)
		}
	case "volume":
		if exists {
			voice.player.SetVolume(float64(command.Volume))
		}
	case "pan":
		if exists {
			voice.pan.setPan(command.Pan)
		}
	case "stop":
		if exists {
			b.stop(command.ID, voice)
		}
	case "pcm-open", "pcm-write":
		// shortcut: cinematic PCM retains the muted adapter behavior until its playback clock is implemented.
	default:
		return fmt.Errorf("ebitengine audio: unknown command %q", command.Kind)
	}
	return nil
}

func (b *ebitAudioBackend) play(command audio.Command) error {
	if command.Format != ".wav" {
		return fmt.Errorf("unsupported audio format %q", command.Format)
	}
	if err := validateWAVPayload(command.Data); err != nil {
		return err
	}
	context := ebitaudio.CurrentContext()
	rate := ebitAudioRate
	if context != nil {
		rate = context.SampleRate()
	}

	decoded, length, err := b.soundSource(command, rate)
	if err != nil {
		return err
	}
	var source io.Reader = decoded
	if command.Loop {
		if length == 0 {
			return fmt.Errorf("cannot loop an empty WAV")
		}

		source = ebitaudio.NewInfiniteLoop(decoded, length)
	}
	pan := &stereoPanReader{source: source}
	pan.setPan(command.Pan)
	player, err := b.newPlayer(pan)
	if err != nil {
		return err
	}
	player.SetVolume(float64(command.Volume))
	player.Play()
	b.voices[command.ID] = ebitVoice{player: player, pan: pan}
	return nil
}

// soundSource shares immutable short-effect PCM, never player cursors or controls.
// Content hashes include changed assets without requiring a second reload owner.
func (b *ebitAudioBackend) soundSource(command audio.Command, rate int) (io.ReadSeeker, int64, error) {
	key := ""
	if !command.Stream {
		key = fmt.Sprintf("%d:%x", rate, sha256.Sum256(command.Data))
		if value, ok := b.samples.Retrieve(key); ok {
			sample := value.(*ebitSample)
			return sample.reader(command.Loop), sample.length, nil
		}
	}

	decoded, err := wav.DecodeWithSampleRate(rate, bytes.NewReader(command.Data))
	if err != nil {
		return nil, 0, err
	}

	length := decoded.Length()
	if command.Stream || length <= 0 || length > ebitMaxSampleBytes {
		return decoded, length, nil
	}

	// Reserve fixed blocks up front without resampling on the frame thread.
	// Loop readers expose only the authored length, not padding used by one-shots.
	capacity := (int(length)/4096 + 1) * 4096
	sample := &ebitSample{decoder: decoded, data: make([]byte, 0, capacity), length: length}
	// Charge encoded input and a conservative decoder reservation until eviction,
	// even though completing the sample releases both before its PCM is evicted.
	weight := capacity + len(command.Data) + len(key) + ebitDecoderReserveBytes + 128
	if err := b.samples.Insert(key, sample, weight); err != nil {
		b.logger.Error("retaining Ebitengine samples", "error", err)
	}

	return sample.reader(command.Loop), length, nil
}

// validateWAVPayload bounds decoder allocations and rejects incomplete frames before a loop can retry EOF forever.
func validateWAVPayload(data []byte) error {
	if len(data) < 12 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return fmt.Errorf("invalid WAV header")
	}
	end := int64(binary.LittleEndian.Uint32(data[4:])) + 8
	if end < 12 || end > int64(len(data)) {
		return fmt.Errorf("incomplete WAV container")
	}
	frameBytes := int64(0)
	for offset := int64(12); offset+8 <= end; {
		size := int64(binary.LittleEndian.Uint32(data[offset+4:]))
		next := offset + 8 + size
		if next > end {
			return fmt.Errorf("incomplete WAV chunk")
		}
		kind := string(data[offset : offset+4])
		// shortcut: reject odd header chunks until the pinned WAV decoder handles RIFF padding.
		if kind != "data" && size%2 != 0 {
			return fmt.Errorf("odd WAV header chunks are unsupported")
		}
		switch kind {
		case "fmt ":
			if size < 16 || binary.LittleEndian.Uint32(data[offset+12:]) == 0 {
				return fmt.Errorf("invalid WAV format or sample rate")
			}
			channels := binary.LittleEndian.Uint16(data[offset+10:])
			bits := binary.LittleEndian.Uint16(data[offset+22:])
			frameBytes = int64(channels) * int64(bits/8)
		case "data":
			if frameBytes == 0 || size%frameBytes != 0 {
				return fmt.Errorf("incomplete WAV sample frame")
			}
			return nil
		}
		offset = next + size%2
	}
	return fmt.Errorf("WAV data chunk is missing")
}

func (b *ebitAudioBackend) stop(id audio.SoundID, voice ebitVoice) {
	if err := voice.player.Close(); err != nil {
		b.logger.Error("closing Ebitengine sound", "sound", id, "error", err)
	}
	delete(b.voices, id)
}

// update releases mixer slots outside Drain, which holds the mixer lock while applying commands.
func (b *ebitAudioBackend) update() {
	for id, voice := range b.voices {
		if !voice.player.IsPlaying() {
			b.stop(id, voice)
			b.finished = append(b.finished, id)
		}
	}
	for _, id := range b.finished {
		b.mixer.Release(id)
	}
	b.finished = nil
}

func (b *ebitAudioBackend) close() {
	for id, voice := range b.voices {
		b.stop(id, voice)
		b.mixer.Release(id)
	}
	for _, id := range b.finished {
		b.mixer.Release(id)
	}
	b.finished = nil
	b.samples.Clear()
}
