package audio

import (
	"reflect"
	"testing"
	"time"
)

type recordingBackend struct{ commands []Command }

func (b *recordingBackend) Apply(command Command) error {
	b.commands = append(b.commands, command)
	return nil
}

func TestPCMStreamUsesCheckedAudioOwnership(t *testing.T) {
	var mixer Mixer
	id, err := mixer.OpenPCMStream(44100, 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := mixer.WritePCM(id, []byte{1, 0, 2, 0}); err != nil {
		t.Fatal(err)
	}
	if err := mixer.Stop(id); err != nil {
		t.Fatal(err)
	}
	backend := &recordingBackend{}
	if err := mixer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	if len(backend.commands) != 3 || backend.commands[0].Kind != "pcm-open" || backend.commands[1].Kind != "pcm-write" || backend.commands[2].Kind != "stop" {
		t.Fatalf("PCM commands = %#v", backend.commands)
	}
	if backend.commands[0].Rate != 44100 || backend.commands[0].Channels != 2 {
		t.Fatalf("PCM format = %#v", backend.commands[0])
	}
	if err := mixer.WritePCM(id, []byte{0, 0}); err == nil {
		t.Fatal("wrote PCM through stale stream handle")
	}
}

func TestPCMStreamReportsGenerationCheckedPlaybackClock(t *testing.T) {
	var mixer Mixer
	id, err := mixer.OpenPCMStream(48000, 2)
	if err != nil {
		t.Fatal(err)
	}
	if _, available := mixer.PCMTime(id); available {
		t.Fatal("PCM clock was available before the device consumed frames")
	}
	if err := mixer.ReportPCMFrames(id, 4800); err != nil {
		t.Fatal(err)
	}
	if elapsed, available := mixer.PCMTime(id); !available || elapsed != 100*time.Millisecond {
		t.Fatalf("PCM clock = %v, %v", elapsed, available)
	}
	if err := mixer.Stop(id); err != nil {
		t.Fatal(err)
	}
	if _, available := mixer.PCMTime(id); available {
		t.Fatal("stale PCM clock remained available")
	}
	if err := mixer.ReportPCMFrames(id, 1); err == nil {
		t.Fatal("stale PCM progress was accepted")
	}
}

func TestMixerDeterministicFadeAndGroupStop(t *testing.T) {
	t.Parallel()
	var mixer Mixer
	id, err := mixer.PlayWithOptions(".wav", []byte("wave"), PlayOptions{Bus: "ambience", Volume: 1, Group: "rain"})
	if err != nil {
		t.Fatal(err)
	}
	if err := mixer.Fade(id, 0, 100*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	mixer.Advance(25 * time.Millisecond)
	mixer.Advance(75 * time.Millisecond)
	if err := mixer.StopGroup("rain"); err != nil {
		t.Fatal(err)
	}
	backend := &recordingBackend{}
	if err := mixer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	if len(backend.commands) != 4 || backend.commands[1].Volume != .75 || backend.commands[2].Volume != 0 || backend.commands[3].Kind != "stop" {
		t.Fatalf("commands = %#v", backend.commands)
	}
}

func TestMixerQueuesCheckedSoundLifetime(t *testing.T) {
	t.Parallel()

	var mixer Mixer
	id, err := mixer.Play(".wav", []byte("wave"))
	if err != nil {
		t.Fatal(err)
	}
	if err := mixer.SetVolume(id, .5); err != nil {
		t.Fatal(err)
	}
	if err := mixer.Stop(id); err != nil {
		t.Fatal(err)
	}
	if err := mixer.SetVolume(id, 1); err == nil {
		t.Fatal("expected stale handle to fail")
	}
	backend := &recordingBackend{}
	if err := mixer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	var kinds []string
	for _, command := range backend.commands {
		kinds = append(kinds, command.Kind)
	}
	if !reflect.DeepEqual(kinds, []string{"play", "volume", "stop"}) {
		t.Fatalf("commands = %v", kinds)
	}
	replacement, err := mixer.Play(".wav", []byte("wave"))
	if err != nil {
		t.Fatal(err)
	}
	if replacement.Slot != id.Slot || replacement.Generation == id.Generation {
		t.Fatalf("replacement = %#v, original = %#v", replacement, id)
	}
}

func TestMixerRoutesBusVolumePanAndLoop(t *testing.T) {
	t.Parallel()
	var mixer Mixer
	if err := mixer.SetBusVolume("music", .5); err != nil {
		t.Fatal(err)
	}
	id, err := mixer.PlayWithOptions(".wav", []byte("wave"), PlayOptions{Bus: "music", Volume: .8, Pan: -.5, Loop: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := mixer.SetBusVolume("music", .25); err != nil {
		t.Fatal(err)
	}
	if err := mixer.SetPan(id, 1); err != nil {
		t.Fatal(err)
	}
	backend := &recordingBackend{}
	if err := mixer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	if len(backend.commands) != 3 || backend.commands[0].Volume != .4 || !backend.commands[0].Loop || backend.commands[1].Volume != .2 || backend.commands[2].Pan != 1 {
		t.Fatalf("commands = %#v", backend.commands)
	}
}

func TestMixerPreservesStreamingIntent(t *testing.T) {
	t.Parallel()
	var mixer Mixer
	if _, err := mixer.PlayWithOptions(".wav", []byte("music"), PlayOptions{Bus: "music", Volume: 1, Stream: true, Loop: true}); err != nil {
		t.Fatal(err)
	}
	backend := &recordingBackend{}
	if err := mixer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	if len(backend.commands) != 1 || !backend.commands[0].Stream || !backend.commands[0].Loop {
		t.Fatalf("commands = %#v", backend.commands)
	}
}
