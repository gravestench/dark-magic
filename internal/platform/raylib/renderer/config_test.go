package raylibRenderer

import (
	"encoding/json"
	"reflect"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestDefaultConfigDataRoundTripsCompleteConfiguration(t *testing.T) {
	want := DefaultConfig()
	service := &Service{}
	var got Config
	if err := json.Unmarshal(service.DefaultConfigData(), &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("config = %#v, want %#v", got, want)
	}
	if got.Window.Width <= 0 || got.Window.Height <= 0 || got.Resolution.Width <= 0 || got.Resolution.Height <= 0 || got.Cache.BudgetMB <= 0 {
		t.Fatal("default configuration is incomplete")
	}
}

func TestWindowConfigFlagsUseDesktopBorderlessMode(t *testing.T) {
	config := DefaultConfig()
	config.Window.Borderless = true
	flags := windowConfigFlags(config)
	for _, required := range []uint32{rl.FlagWindowResizable, rl.FlagBorderlessWindowedMode, rl.FlagWindowMaximized} {
		if flags&required == 0 {
			t.Fatalf("flags %#x omit %#x", flags, required)
		}
	}
	if flags&rl.FlagFullscreenMode != 0 {
		t.Fatalf("borderless desktop mode unexpectedly requests exclusive fullscreen: %#x", flags)
	}
}
