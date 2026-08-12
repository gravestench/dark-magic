package d2legacy

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

type audioCommands struct{ values []audio.Command }

func (commands *audioCommands) Apply(command audio.Command) error {
	commands.values = append(commands.values, command)
	return nil
}

func TestLuaOwnsLegacySoundRecordSelection(t *testing.T) {
	assets := fstest.MapFS{
		"data/global/excel/Sounds.txt": {Data: []byte("Sound\tFileName\tIsUI\tVolume Min\tVolume Max\tGroup Size\tGroup Weight\nmenu\tui/one.wav\t1\t255\t255\t2\t1\nmenu_alt\tui/two.wav\t1\t255\t255\t0\t3\n")},
		"data/global/sfx/ui/one.wav":   {Data: []byte("one")},
		"data/global/sfx/ui/two.wav":   {Data: []byte("two")},
	}
	source, err := content.New(content.Layer{Name: "test", FS: assets}, content.Layer{Name: "d2legacy", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}
	var mixer audio.Mixer
	runtime := modruntime.New()
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(source, "lua")); err != nil {
		t.Fatal(err)
	}
	for _, module := range []modruntime.Module{modruntime.AudioModule(runtime, &mixer, source), modruntime.RecordsModule(recordstore.New(source))} {
		if err := runtime.RegisterModule(module); err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	scope := &modruntime.Scope{}
	defer scope.Close()
	if err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error {
		return state.DoString(`require("d2legacy.audio").play_record("menu", 2)`)
	}); err != nil {
		t.Fatal(err)
	}
	backend := &audioCommands{}
	if err := mixer.Drain(backend); err != nil {
		t.Fatal(err)
	}
	if len(backend.values) != 1 || backend.values[0].Volume != 1 || string(backend.values[0].Data) != "two" {
		t.Fatalf("Lua-selected playback = %#v", backend.values)
	}
}
