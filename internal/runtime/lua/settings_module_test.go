package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/preferences"
)

func TestSettingsModuleAppliesAudioPreferences(t *testing.T) {
	runtime := New()
	settings := preferences.NewTransient()
	var mixer audio.Mixer
	if err := runtime.RegisterModule(SettingsModule(settings, &mixer)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	scripts := fstest.MapFS{"script.lua": &fstest.MapFile{Data: []byte(`local s=require("dm.settings/v1"); s.set("sound_volume",.25); s.set("music_volume",.75)`)}}
	if err := runtime.Execute(context.Background(), scripts, "script.lua"); err != nil {
		t.Fatal(err)
	}
	diagnostics := mixer.Diagnostics()
	if diagnostics.BusVolumes["sfx"] != .25 || diagnostics.BusVolumes["speech"] != .25 || diagnostics.BusVolumes["music"] != .75 {
		t.Fatalf("bus volumes = %#v", diagnostics.BusVolumes)
	}
}
