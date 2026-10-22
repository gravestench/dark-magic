//go:build ebitengine

package desktop

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/cache"
)

type testAudioPlayer struct {
	source  io.Reader
	playing bool
	volume  float64
	closed  int
}

func (p *testAudioPlayer) Play()               { p.playing = true }
func (p *testAudioPlayer) IsPlaying() bool     { return p.playing }
func (p *testAudioPlayer) SetVolume(v float64) { p.volume = v }
func (p *testAudioPlayer) Close() error        { p.closed++; p.playing = false; return nil }

func testWAV() []byte {
	var result bytes.Buffer
	result.WriteString("RIFF")
	_ = binary.Write(&result, binary.LittleEndian, uint32(40))
	result.WriteString("WAVEfmt ")
	for _, v := range []any{uint32(16), uint16(1), uint16(2), uint32(48000), uint32(192000), uint16(4), uint16(16)} {
		_ = binary.Write(&result, binary.LittleEndian, v)
	}
	result.WriteString("data")
	_ = binary.Write(&result, binary.LittleEndian, uint32(4))
	result.Write([]byte{0xe8, 3, 0xe8, 3})
	return result.Bytes()
}

func testEbitAudio(mixer *audio.Mixer) (*ebitAudioBackend, *[]*testAudioPlayer) {
	backend := newEbitAudioBackend(mixer, slog.New(slog.NewTextHandler(io.Discard, nil)))
	players := []*testAudioPlayer{}
	backend.newPlayer = func(source io.Reader) (ebitAudioPlayer, error) {
		player := &testAudioPlayer{source: source}
		players = append(players, player)
		return player, nil
	}
	return backend, &players
}

func TestEbitAudioControlsAndLoop(t *testing.T) {
	var mixer audio.Mixer
	backend, players := testEbitAudio(&mixer)
	id, err := mixer.PlayWithOptions(".wav", testWAV(), audio.PlayOptions{Bus: "music", Volume: 1, Loop: true, Stream: true, Group: "frontend"})
	if err != nil {
		t.Fatal(err)
	}
	if err := mixer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	player := (*players)[0]
	if !player.playing {
		t.Fatal("player did not start")
	}
	if err := mixer.SetPan(id, 1); err != nil {
		t.Fatal(err)
	}
	if err := mixer.Fade(id, 0, 100*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	mixer.Advance(50 * time.Millisecond)
	if err := mixer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	if player.volume != .5 {
		t.Fatalf("volume %v", player.volume)
	}
	data := make([]byte, 16)
	if _, err := io.ReadFull(player.source, data); err != nil {
		t.Fatal(err)
	}
	for offset := 0; offset < len(data); offset += 4 {
		if binary.LittleEndian.Uint16(data[offset:]) != 0 || binary.LittleEndian.Uint16(data[offset+2:]) != 1000 {
			t.Fatalf("loop/pan output %v", data)
		}
	}
	if err := mixer.StopGroup("frontend"); err != nil {
		t.Fatal(err)
	}
	if err := mixer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	backend.close()
	if player.closed != 1 || len(backend.voices) != 0 {
		t.Fatalf("close count=%d voices=%d", player.closed, len(backend.voices))
	}
}

func TestEbitAudioReusesShortPCMWithIndependentVoices(t *testing.T) {
	var mixer audio.Mixer

	backend, players := testEbitAudio(&mixer)
	for _, pan := range []float32{-1, 1} {
		if _, err := mixer.PlayWithOptions(".wav", testWAV(), audio.PlayOptions{Volume: 1, Pan: pan}); err != nil {
			t.Fatal(err)
		}
	}

	if err := mixer.Drain(backend); err != nil {
		t.Fatal(err)
	}

	var shared *ebitSample

	for index, player := range *players {
		pan := player.source.(*stereoPanReader)

		reader, ok := pan.source.(*ebitSampleReader)
		if !ok {
			t.Fatalf("short voice %d does not share decoded samples: %T", index, pan.source)
		}

		if index == 0 {
			shared = reader.sample
			if len(shared.data) != 0 {
				t.Fatal("play decoded samples on the frame thread")
			}
		} else if shared != reader.sample || shared.decoder != nil {
			t.Fatal("voice did not reuse completed sample")
		}

		data, err := io.ReadAll(player.source)
		if err != nil || len(data) != 4 {
			t.Fatalf("voice %d: data=%v err=%v", index, data, err)
		}

		left, right := binary.LittleEndian.Uint16(data), binary.LittleEndian.Uint16(data[2:])
		if index == 0 && (left != 1000 || right != 0) || index == 1 && (left != 0 || right != 1000) {
			t.Fatalf("voice %d: left=%d right=%d", index, left, right)
		}
	}

	stats := backend.samples.Diagnostics()
	if stats.Entries != 1 || stats.Hits != 1 || stats.Misses != 1 {
		t.Fatalf("cache %+v", stats)
	}

	backend.close()

	if backend.samples.Diagnostics().Weight != 0 {
		t.Fatal("shutdown retained decoded samples")
	}
}

func TestEbitAudioSharedSampleReadersAreIndependent(t *testing.T) {
	var mixer audio.Mixer

	backend, _ := testEbitAudio(&mixer)
	defer backend.close()

	data := testWAV()
	for i := 0; i < 10000; i++ {
		data = append(data, byte(i), byte(i>>8), byte(i), byte(i>>8))
	}

	binary.LittleEndian.PutUint32(data[4:], uint32(len(data)-8))
	binary.LittleEndian.PutUint32(data[40:], uint32(len(data)-44))
	binary.LittleEndian.PutUint32(data[24:], 22050)
	binary.LittleEndian.PutUint32(data[28:], 22050*4)

	first, _, err := backend.soundSource(audio.Command{Data: data}, 48000)
	if err != nil {
		t.Fatal(err)
	}

	second, _, err := backend.soundSource(audio.Command{Data: data}, 48000)
	if err != nil {
		t.Fatal(err)
	}

	var group sync.WaitGroup

	outputs := make([][]byte, 2)

	for i, source := range []io.Reader{first, second} {
		group.Add(1)
		go func() {
			defer group.Done()

			var err error

			outputs[i], err = io.ReadAll(source)
			if err != nil {
				t.Error(err)
			}
		}()
	}

	group.Wait()

	if len(outputs[0]) == 0 || !bytes.Equal(outputs[0], outputs[1]) {
		t.Fatal("concurrent voices differ")
	}

	if first.(*ebitSampleReader).sample.decoder != nil {
		t.Fatal("completed sample retained decoder and encoded input")
	}
}

func TestEbitAudioSampleCacheEvictionDoesNotStopLiveVoice(t *testing.T) {
	var mixer audio.Mixer

	backend, players := testEbitAudio(&mixer)
	backend.samples = cache.New(ebitDecoderReserveBytes + 5000)

	for _, value := range []byte{1, 2, 1} {
		data := testWAV()

		data[44] = value
		if _, err := mixer.Play(".wav", data); err != nil {
			t.Fatal(err)
		}
	}

	if err := mixer.Drain(backend); err != nil {
		t.Fatal(err)
	}

	stats := backend.samples.Diagnostics()
	if stats.Entries != 1 || stats.Weight > stats.Budget || stats.Evictions != 2 || stats.Misses != 3 {
		t.Fatalf("cache %+v", stats)
	}

	for index, player := range *players {
		data, err := io.ReadAll(player.source)

		want := byte(1)
		if index == 1 {
			want = 2
		}

		if err != nil || len(data) != 4 || data[0] != want {
			t.Fatalf("voice %d after eviction: data=%v err=%v", index, data, err)
		}
	}

	backend.close()
}

func TestEbitAudioStreamingAndOversizedSoundsBypassSampleCache(t *testing.T) {
	large := testWAV()
	large = append(large, make([]byte, 1024*1024)...)
	binary.LittleEndian.PutUint32(large[4:], uint32(len(large)-8))
	binary.LittleEndian.PutUint32(large[40:], uint32(len(large)-44))

	for _, c := range []struct {
		data   []byte
		stream bool
	}{{testWAV(), true}, {large, false}} {
		var mixer audio.Mixer

		backend, players := testEbitAudio(&mixer)
		if _, err := mixer.PlayWithOptions(".wav", c.data, audio.PlayOptions{Volume: 1, Stream: c.stream}); err != nil {
			t.Fatal(err)
		}

		if err := mixer.Drain(backend); err != nil {
			t.Fatal(err)
		}

		if _, ok := (*players)[0].source.(*stereoPanReader).source.(*ebitSampleReader); ok {
			t.Fatal("stream buffered in full")
		}

		if backend.samples.Diagnostics().Entries != 0 {
			t.Fatal("stream entered sample cache")
		}

		backend.close()
	}
}

func TestEbitAudioCachedResamplingMatchesStreaming(t *testing.T) {
	data := testWAV()
	for i := 0; i < 1000; i++ {
		data = append(data, byte(i), byte(i>>8), byte(1000-i), byte((1000-i)>>8))
	}

	binary.LittleEndian.PutUint32(data[4:], uint32(len(data)-8))
	binary.LittleEndian.PutUint32(data[40:], uint32(len(data)-44))
	binary.LittleEndian.PutUint32(data[24:], 22050)
	binary.LittleEndian.PutUint32(data[28:], 22050*4)

	for _, loop := range []bool{false, true} {
		var mixer audio.Mixer

		backend, players := testEbitAudio(&mixer)

		for _, stream := range []bool{true, false, false} {
			options := audio.PlayOptions{Volume: 1, Loop: loop, Stream: stream}
			if _, err := mixer.PlayWithOptions(".wav", data, options); err != nil {
				t.Fatal(err)
			}

			if err := mixer.Drain(backend); err != nil {
				t.Fatal(err)
			}

			backend.update()
		}

		var reference []byte

		for index, player := range *players {
			output, err := io.ReadAll(io.LimitReader(player.source, 32768))
			if err != nil {
				t.Fatal(err)
			}

			if index == 0 {
				reference = output
			} else if !bytes.Equal(reference, output) {
				t.Fatalf("loop=%v voice=%d differs from streaming", loop, index)
			}
		}

		if backend.samples.Diagnostics().Hits != 1 {
			t.Fatal("frame update lost cached samples")
		}

		backend.close()
	}
}

func TestEbitAudioSampleCacheSeparatesOutputRates(t *testing.T) {
	var mixer audio.Mixer

	backend, _ := testEbitAudio(&mixer)
	defer backend.close()

	data := testWAV()
	data = append(data, make([]byte, 4000)...)
	binary.LittleEndian.PutUint32(data[4:], uint32(len(data)-8))
	binary.LittleEndian.PutUint32(data[40:], uint32(len(data)-44))

	command := audio.Command{Data: data}
	for _, rate := range []int{44100, 48000, 44100} {
		source, length, err := backend.soundSource(command, rate)
		if err != nil {
			t.Fatal(err)
		}

		want := int64(len(data)-44) * int64(rate) / 48000 / 4 * 4
		if length != want {
			t.Fatalf("rate=%d length=%d want=%d", rate, length, want)
		}

		if _, err := io.ReadAll(source); err != nil {
			t.Fatal(err)
		}
	}

	stats := backend.samples.Diagnostics()
	if stats.Entries != 2 || stats.Hits != 1 {
		t.Fatalf("rate-separated cache %+v", stats)
	}
}

func BenchmarkEbitAudioSource(b *testing.B) {
	for _, seconds := range []int{1, 5} {
		for _, cold := range []bool{true, false} {
			name := fmt.Sprintf("%ds/cold=%v", seconds, cold)
			b.Run(name, func(b *testing.B) {
				var mixer audio.Mixer

				backend, _ := testEbitAudio(&mixer)
				defer backend.close()

				data := testWAV()
				data = append(data, make([]byte, seconds*22050*4)...)
				binary.LittleEndian.PutUint32(data[4:], uint32(len(data)-8))
				binary.LittleEndian.PutUint32(data[40:], uint32(len(data)-44))
				binary.LittleEndian.PutUint32(data[24:], 22050)
				binary.LittleEndian.PutUint32(data[28:], 22050*4)

				command := audio.Command{Data: data}
				if _, _, err := backend.soundSource(command, 48000); err != nil {
					b.Fatal(err)
				}

				b.ReportAllocs()
				b.ResetTimer()

				for range b.N {
					if cold {
						backend.samples.Clear()
					}

					if _, _, err := backend.soundSource(command, 48000); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func TestEbitAudioCompletionIsBoundedAndBadAssetDoesNotBlock(t *testing.T) {
	var mixer audio.Mixer
	backend, players := testEbitAudio(&mixer)
	if _, err := mixer.Play(".wav", []byte("broken")); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1000; i++ {
		if _, err := mixer.Play(".wav", testWAV()); err != nil {
			t.Fatal(err)
		}
		if err := mixer.Drain(backend); err != nil {
			t.Fatal(err)
		}
		player := (*players)[len(*players)-1]
		if data, err := io.ReadAll(player.source); err != nil || len(data) != 4 {
			t.Fatalf("decoded length=%d err=%v", len(data), err)
		}
		player.playing = false
		backend.update()
		if err := mixer.Drain(backend); err != nil {
			t.Fatal(err)
		}
		if player.closed != 1 {
			t.Fatalf("close count %d", player.closed)
		}
	}
	stats := mixer.Diagnostics()
	if stats.Active != 0 || stats.Pending != 0 || stats.Slots > 2 || len(backend.voices) != 0 {
		t.Fatalf("stats %+v voices=%d", stats, len(backend.voices))
	}
}

func TestEbitAudioRejectsTruncatedLoopsBeforeCreatingPlayer(t *testing.T) {
	truncated := testWAV()[:44]
	oversizedChunk := testWAV()
	binary.LittleEndian.PutUint32(oversizedChunk[16:], ^uint32(0))
	zeroRate := testWAV()
	binary.LittleEndian.PutUint32(zeroRate[24:], 0)
	partialFrame := testWAV()[:47]
	binary.LittleEndian.PutUint32(partialFrame[4:], 39)
	binary.LittleEndian.PutUint32(partialFrame[40:], 3)
	padded := append([]byte{}, testWAV()[:12]...)
	padded = append(padded, 'J', 'U', 'N', 'K', 1, 0, 0, 0, 0, 0)
	padded = append(padded, testWAV()[12:]...)
	binary.LittleEndian.PutUint32(padded[4:], uint32(len(padded)-8))
	for _, bad := range [][]byte{truncated, oversizedChunk, zeroRate, partialFrame, padded} {
		var mixer audio.Mixer
		backend, players := testEbitAudio(&mixer)
		if _, err := mixer.PlayWithOptions(".wav", bad, audio.PlayOptions{Volume: 1, Loop: true}); err != nil {
			t.Fatal(err)
		}
		if _, err := mixer.Play(".wav", testWAV()); err != nil {
			t.Fatal(err)
		}
		if err := mixer.Drain(backend); err != nil {
			t.Fatal(err)
		}
		if len(*players) != 1 {
			t.Fatalf("invalid WAV reached player: count=%d", len(*players))
		}
		backend.update()
		if mixer.Diagnostics().Active != 1 {
			t.Fatal("failed sound was not retired")
		}
		backend.close()
	}
}

func TestEbitAudioPCMDoesNotBlockSoundsAndCloseReleases(t *testing.T) {
	var mixer audio.Mixer
	backend, _ := testEbitAudio(&mixer)
	pcm, err := mixer.OpenPCMStream(22050, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := mixer.WritePCM(pcm, []byte{0, 0}); err != nil {
		t.Fatal(err)
	}
	id, err := mixer.Play(".wav", testWAV())
	if err != nil {
		t.Fatal(err)
	}
	if err := mixer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	if len(backend.voices) != 1 {
		t.Fatalf("voices %d", len(backend.voices))
	}
	if _, ok := mixer.PCMTime(pcm); ok {
		t.Fatal("muted PCM reported a device clock")
	}
	backend.close()
	if mixer.Active(id) {
		t.Fatal("closed sound remains active")
	}
	if err := mixer.Stop(pcm); err != nil {
		t.Fatal(err)
	}
	if err := mixer.Drain(backend); err != nil {
		t.Fatal(err)
	}
}

func TestEbitAudioAttachmentDiagnostics(t *testing.T) {
	for _, native := range []bool{false, true} {
		name, expected := "muted", "muted"
		if native {
			name, expected = "native", "native"
		}
		t.Run(name, func(t *testing.T) {
			renderer := newEbitRenderer(Options{NativeAudio: native, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
			before := renderer.BackendDiagnostics().(EbitengineDiagnostics)
			if before.Audio != "muted" {
				t.Fatalf("unattached audio = %q", before.Audio)
			}
			if err := renderer.AttachAudio(nil); err == nil {
				t.Fatal("nil mixer accepted")
			}
			if got := renderer.BackendDiagnostics().(EbitengineDiagnostics).Audio; got != "muted" {
				t.Fatalf("failed attachment audio = %q", got)
			}
			var readers sync.WaitGroup
			readers.Add(1)
			go func() {
				defer readers.Done()
				for range 1000 {
					got := renderer.BackendDiagnostics().(EbitengineDiagnostics).Audio
					if got != "muted" && got != "native" {
						t.Errorf("concurrent audio diagnostics = %q", got)
						return
					}
				}
			}()
			defer readers.Wait()
			if err := renderer.AttachAudio(&audio.Mixer{}); err != nil {
				t.Fatal(err)
			}
			if got := renderer.BackendDiagnostics().(EbitengineDiagnostics).Audio; got != expected {
				t.Fatalf("attached audio = %q, want %q", got, expected)
			}
			if (renderer.audio != nil) != native {
				t.Fatal("diagnostics disagree with attached native adapter")
			}
			if err := renderer.AttachAudio(&audio.Mixer{}); err == nil {
				t.Fatal("duplicate mixer accepted")
			}
			if got := renderer.BackendDiagnostics().(EbitengineDiagnostics).Audio; got != expected {
				t.Fatalf("rejected duplicate changed audio to %q", got)
			}
			if before.Audio != "muted" {
				t.Fatal("previous diagnostics snapshot changed")
			}
			renderer.started.Store(true)
			if err := renderer.Stop(context.Background()); err != nil {
				t.Fatal(err)
			}
			if got := renderer.BackendDiagnostics().(EbitengineDiagnostics).Audio; got != "muted" {
				t.Fatalf("stopped audio = %q", got)
			}
		})
	}
}
