package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/audiocore"
	"github.com/gravestench/dark-magic/internal/host"
)

func TestAudioHandlesBelongToLuaComponentScope(t *testing.T) {
	t.Parallel()

	var mixer audiocore.Mixer
	source := fstest.MapFS{
		"sound.wav": &fstest.MapFile{Data: []byte("wave")},
		"system.lua": &fstest.MapFile{Data: []byte(`
local audio = require("dm.audio/v1")
return { id = "sound", start = function(self) self.sound = audio.play("sound.wav"); self.sound:set_volume(.5) end }
`)},
	}
	runtime := New()
	if err := runtime.RegisterModule(AudioModule(runtime, &mixer, source)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	definition, err := LoadDefinition(context.Background(), runtime, source, "system.lua")
	if err != nil {
		t.Fatal(err)
	}
	manager := host.NewManager()
	if err := manager.Register(definition.Managed()); err != nil {
		t.Fatal(err)
	}
	if err := manager.Enable(context.Background(), definition.ID); err != nil {
		t.Fatal(err)
	}
	backend := &recordingAudioBackend{}
	if err := mixer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	if err := manager.Disable(context.Background(), definition.ID); err != nil {
		t.Fatal(err)
	}
	if err := mixer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	if len(backend.commands) != 3 || backend.commands[2].Kind != "stop" {
		t.Fatalf("commands = %#v", backend.commands)
	}
}

type recordingAudioBackend struct{ commands []audiocore.Command }

func (b *recordingAudioBackend) Apply(command audiocore.Command) error {
	b.commands = append(b.commands, command)
	return nil
}
