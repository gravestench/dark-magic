package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/preferences"
)

type recordingRenderSettings struct {
	debug  bool
	budget uint64
	cache  uint64
}

func (r *recordingRenderSettings) SetResidencyDebug(value bool)        { r.debug = value }
func (r *recordingRenderSettings) SetTextureUploadBudget(value uint64) { r.budget = value }
func (r *recordingRenderSettings) SetTextureCacheBudget(value uint64)  { r.cache = value }

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
	scripts := fstest.MapFS{"script.lua": &fstest.MapFile{Data: []byte(`local s=require("engine.settings/v1"); s.set("sound_volume",.25); s.set("music_volume",.75)`)}}
	if err := runtime.Execute(context.Background(), scripts, "script.lua"); err != nil {
		t.Fatal(err)
	}
	diagnostics := mixer.Diagnostics()
	if diagnostics.BusVolumes["sfx"] != .25 || diagnostics.BusVolumes["speech"] != .25 || diagnostics.BusVolumes["music"] != .75 {
		t.Fatalf("bus volumes = %#v", diagnostics.BusVolumes)
	}
}

func TestSettingsModuleAppliesPersistentRenderDiagnosticsPreferences(t *testing.T) {
	runtime := New()
	settings, err := preferences.New(t.TempDir() + "/preferences.json")
	if err != nil {
		t.Fatal(err)
	}
	var mixer audio.Mixer
	target := &recordingRenderSettings{}
	if err := runtime.RegisterModule(SettingsModule(settings, &mixer, target)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	script := `local s=require("engine.settings/v1"); s.set("debug_texture_residency",true); s.set("texture_upload_budget_mb",8); s.set("texture_cache_budget_mb",768); s.save(); assert(s.get("debug_texture_residency") and s.get("texture_upload_budget_mb")==8 and s.get("texture_cache_budget_mb")==768)`
	if err := runtime.Execute(context.Background(), fstest.MapFS{"script.lua": {Data: []byte(script)}}, "script.lua"); err != nil {
		t.Fatal(err)
	}
	if !target.debug || target.budget != 8*1024*1024 || target.cache != 768*1024*1024 {
		t.Fatalf("render settings = %#v", target)
	}
	reloaded, err := preferences.New(settings.Path())
	if err != nil {
		t.Fatal(err)
	}
	if values := reloaded.Values(); !values.DebugTextureResidency || values.TextureUploadBudgetMB != 8 || values.TextureCacheBudgetMB != 768 {
		t.Fatalf("persisted values = %#v", values)
	}
}
