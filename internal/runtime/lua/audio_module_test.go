package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/game/data/catalog"
	"github.com/gravestench/dark-magic/internal/game/data/store"
)

func TestAudioHandlesBelongToLuaComponentScope(t *testing.T) {
	t.Parallel()

	var mixer audio.Mixer
	source := fstest.MapFS{
		"sound.wav": &fstest.MapFile{Data: []byte("wave")},
		"system.lua": &fstest.MapFile{Data: []byte(`
local audio = require("engine.audio/v1")
return { id = "sound", start = function(self) audio.set_bus_volume("ui", .8); self.sound = audio.play("sound.wav", {bus="ui", volume=.5, pan=-.25, loop=true}); self.sound:set_pan(.25); self.sound:set_volume(.4) end }
`)},
	}
	runtime := New()
	if err := runtime.RegisterModule(AudioModule(runtime, &mixer, source, gamedata.New(recordstore.New(source)))); err != nil {
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
	if len(backend.commands) != 4 || backend.commands[0].Kind != "play" || backend.commands[0].Volume != .4 || !backend.commands[0].Loop || backend.commands[1].Kind != "pan" || backend.commands[3].Kind != "stop" {
		t.Fatalf("commands = %#v", backend.commands)
	}
}

func TestPersistentAudioOutlivesCallingComponent(t *testing.T) {
	var mixer audio.Mixer
	source := fstest.MapFS{
		"music.wav": &fstest.MapFile{Data: []byte("music")},
		"system.lua": &fstest.MapFile{Data: []byte(`
local audio = require("engine.audio/v1")
return { id = "music", start = function(self)
  audio.play_persistent("music.wav", {bus="music", loop=true, stream=true, group="frontend"})
end }
`)},
	}
	runtime := New()
	if err := runtime.RegisterModule(AudioModule(runtime, &mixer, source, gamedata.New(recordstore.New(source)))); err != nil {
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
	if len(backend.commands) != 1 || backend.commands[0].Kind != "play" || !backend.commands[0].Loop || !backend.commands[0].Stream {
		t.Fatalf("persistent commands = %#v", backend.commands)
	}
	if err := mixer.StopGroup("frontend"); err != nil {
		t.Fatal(err)
	}
	if err := mixer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	if len(backend.commands) != 2 || backend.commands[1].Kind != "stop" {
		t.Fatalf("stopped commands = %#v", backend.commands)
	}
}

type recordingAudioBackend struct{ commands []audio.Command }

func (b *recordingAudioBackend) Apply(command audio.Command) error {
	b.commands = append(b.commands, command)
	return nil
}
