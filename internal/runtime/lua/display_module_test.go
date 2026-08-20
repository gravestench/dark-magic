package modruntime

import (
	"context"
	"fmt"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// TestDisplayModuleReportsLiveSurfaceSize proves responsive Lua layout observes renderer changes immediately.
func TestDisplayModuleReportsLiveSurfaceSize(t *testing.T) {
	runtime := New()
	width, height := 1920, 1080
	if err := runtime.RegisterModule(DisplayModule(func() (int, int) { return width, height })); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()
	assertSize := func(expectedWidth, expectedHeight int) {
		t.Helper()
		if err := runtime.Run(context.Background(), func(state *lua.LState) error {
			return state.DoString(fmt.Sprintf(`
local display = require("engine.display/v1")
local width, height = display.size()
assert(width == %d and height == %d)
`, expectedWidth, expectedHeight))
		}); err != nil {
			t.Fatal(err)
		}
	}
	assertSize(1920, 1080)
	width, height = 2560, 1440
	assertSize(2560, 1440)
}
