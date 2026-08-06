package audiocore

import (
	"reflect"
	"testing"
)

type recordingBackend struct{ commands []Command }

func (b *recordingBackend) Apply(command Command) error {
	b.commands = append(b.commands, command)
	return nil
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
