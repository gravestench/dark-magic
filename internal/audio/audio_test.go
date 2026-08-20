package audio

import (
	"reflect"
	"testing"
	"time"
)

// recordingBackend captures commands synchronously so tests can inspect the exact backend protocol order.
type recordingBackend struct {
	commands []Command
}

// Apply records one command and never fails, isolating tests from native audio ownership.
func (b *recordingBackend) Apply(command Command) error {
	b.commands = append(b.commands, command)

	return nil
}

// drainRecordedCommands transfers all queued work into a fresh recorder and fails at the contract boundary.
func drainRecordedCommands(t *testing.T, mixer *Mixer) []Command {
	t.Helper()

	backend := &recordingBackend{}
	if err := mixer.Drain(backend); err != nil {
		t.Fatal(err)
	}

	return backend.commands
}

// assertCommandKinds verifies both command count and ordering without hiding the complete command list on failure.
func assertCommandKinds(t *testing.T, commands []Command, want []string) {
	t.Helper()

	got := make([]string, 0, len(commands))
	for _, command := range commands {
		got = append(got, command.Kind)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("command kinds = %v, want %v; commands = %#v", got, want, commands)
	}
}

// TestPCMStreamUsesCheckedAudioOwnership verifies ordered stream commands and stale-handle rejection after retirement.
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

	commands := drainRecordedCommands(t, &mixer)
	assertCommandKinds(t, commands, []string{commandPCMOpen, commandPCMWrite, commandStop})

	if commands[0].Rate != 44100 || commands[0].Channels != 2 {
		t.Fatalf("PCM format = %#v", commands[0])
	}

	// Retirement advances the generation before any subsequent owner-thread work can reuse the slot.
	if err := mixer.WritePCM(id, []byte{0, 0}); err == nil {
		t.Fatal("wrote PCM through stale stream handle")
	}
}

// TestPCMStreamReportsGenerationCheckedPlaybackClock verifies that device progress belongs only to the live generation.
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

// TestMixerDeterministicFadeAndGroupStop verifies caller time and slot order determine the emitted command sequence.
func TestMixerDeterministicFadeAndGroupStop(t *testing.T) {
	t.Parallel()

	var mixer Mixer

	id, err := mixer.PlayWithOptions(
		".wav",
		[]byte("wave"),
		PlayOptions{Bus: "ambience", Volume: 1, Group: "rain"},
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := mixer.Fade(id, 0, 100*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	// Separate advances prove interpolation is based on accumulated mixer time rather than wall-clock timing.
	mixer.Advance(25 * time.Millisecond)
	mixer.Advance(75 * time.Millisecond)

	if err := mixer.StopGroup("rain"); err != nil {
		t.Fatal(err)
	}

	commands := drainRecordedCommands(t, &mixer)
	assertCommandKinds(t, commands, []string{commandPlay, commandVolume, commandVolume, commandStop})

	if commands[1].Volume != .75 {
		t.Fatalf("intermediate fade volume = %v, want .75; commands = %#v", commands[1].Volume, commands)
	}

	if commands[2].Volume != 0 {
		t.Fatalf("final fade volume = %v, want 0; commands = %#v", commands[2].Volume, commands)
	}
}

// TestMixerQueuesCheckedSoundLifetime verifies retirement rejects stale handles and reuses slots with new generations.
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

	commands := drainRecordedCommands(t, &mixer)
	assertCommandKinds(t, commands, []string{commandPlay, commandVolume, commandStop})

	// LIFO reuse should conserve slots while generation advancement keeps the old handle invalid.
	replacement, err := mixer.Play(".wav", []byte("wave"))
	if err != nil {
		t.Fatal(err)
	}

	if replacement.Slot != id.Slot || replacement.Generation == id.Generation {
		t.Fatalf("replacement = %#v, original = %#v", replacement, id)
	}
}

// TestMixerRoutesBusVolumePanAndLoop verifies logical and bus gains compose without losing playback options.
func TestMixerRoutesBusVolumePanAndLoop(t *testing.T) {
	t.Parallel()

	var mixer Mixer
	if err := mixer.SetBusVolume("music", .5); err != nil {
		t.Fatal(err)
	}

	id, err := mixer.PlayWithOptions(
		".wav",
		[]byte("wave"),
		PlayOptions{Bus: "music", Volume: .8, Pan: -.5, Loop: true},
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := mixer.SetBusVolume("music", .25); err != nil {
		t.Fatal(err)
	}

	if err := mixer.SetPan(id, 1); err != nil {
		t.Fatal(err)
	}

	commands := drainRecordedCommands(t, &mixer)
	assertCommandKinds(t, commands, []string{commandPlay, commandVolume, commandPan})

	if commands[0].Volume != .4 || !commands[0].Loop {
		t.Fatalf("initial routed command = %#v, want volume .4 and looping", commands[0])
	}

	if commands[1].Volume != .2 {
		t.Fatalf("updated routed volume = %v, want .2; commands = %#v", commands[1].Volume, commands)
	}

	if commands[2].Pan != 1 {
		t.Fatalf("updated pan = %v, want 1; commands = %#v", commands[2].Pan, commands)
	}
}

// TestMixerPreservesStreamingIntent verifies stream and loop flags cross the renderer-thread boundary unchanged.
func TestMixerPreservesStreamingIntent(t *testing.T) {
	t.Parallel()

	var mixer Mixer

	_, err := mixer.PlayWithOptions(
		".wav",
		[]byte("music"),
		PlayOptions{Bus: "music", Volume: 1, Stream: true, Loop: true},
	)
	if err != nil {
		t.Fatal(err)
	}

	commands := drainRecordedCommands(t, &mixer)
	assertCommandKinds(t, commands, []string{commandPlay})

	if !commands[0].Stream || !commands[0].Loop {
		t.Fatalf("commands = %#v", commands)
	}
}
